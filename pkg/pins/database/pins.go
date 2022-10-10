package database

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/RediSearch/redisearch-go/redisearch"

	"github.com/crossedbot/axis/pkg/pins/models"
)

const (
	PinsIndexName = "pins"
	PinsKeyPrefix = "pin:"
)

var (
	// Regular Expressions
	TokenDelimitersRe = regexp.MustCompile(`(\b|\B)([,.<>{}\[\]\"\':;!@#\$%\^&\*\(\)\-\+=~ ])(\b|\B)`)

	// Errors
	ErrNotFound = errors.New("pin not found")
)

// Pins represents an interface to the Pins database
type Pins interface {
	// Set adds or updates a Pin
	Set(pinStatus models.PinStatus) error

	// Patch patches the fields of a Pin according to the given ID
	Patch(id string, fields map[string]interface{}) error

	// Get returns the Pin status for a given ID
	Get(id string) (models.PinStatus, error)

	// Find returns a list of Pins for the given parameters
	Find(
		cids, statuses []string,
		name string,
		before, after int64,
		match string,
		limit int, offset int,
		sortBy *SortingKey,
		meta models.Info,
	) (models.Pins, error)

	// Delete removes the Pin according to the given ID
	Delete(id string) error
}

// pins represents the implementation of the Pins interface
type pins struct {
	*redisearch.Client
	ctx context.Context
}

// New returns a new Pins for the given context and database address. If drop is
// set, the data is dropped from the database.
func New(ctx context.Context, addr string, drop bool) (Pins, error) {
	ps := &pins{
		Client: redisearch.NewClient(addr, PinsIndexName),
		ctx:    ctx,
	}
	// If set indicated, drop the index
	if drop {
		ps.Drop()
	}
	schema := redisearch.NewSchema(redisearch.DefaultOptions).
		AddField(redisearch.NewTextField("status")).
		AddField(redisearch.NewNumericField("created")).
		AddField(redisearch.NewTextField("cid")).
		AddField(redisearch.NewSortableTextField("name", 1.0)).
		AddField(redisearch.NewTagField("origins")).
		AddField(redisearch.NewTagFieldOptions(
			"meta", redisearch.TagFieldOptions{
				Separator: ',',
				NoIndex:   false,
				Sortable:  true,
			})).
		AddField(redisearch.NewTagField("delegates")).
		AddField(redisearch.NewTagFieldOptions(
			"info", redisearch.TagFieldOptions{
				Separator: ',',
				NoIndex:   false,
				Sortable:  true,
			}))
	indexExists, err := containsIndex(ps.Client, PinsIndexName)
	if err != nil {
		return nil, err
	} else if !indexExists {
		indexDefinition := redisearch.NewIndexDefinition().
			AddPrefix(PinsKeyPrefix)
		err = ps.CreateIndexWithIndexDefinition(schema, indexDefinition)
	}
	return ps, err
}

func (d *pins) Set(pinStatus models.PinStatus) error {
	created, err := strconv.ParseInt(pinStatus.Created, 10, 64)
	if err != nil {
		return err
	}
	index := fmt.Sprintf("%s%s", PinsKeyPrefix, pinStatus.Id)
	doc := redisearch.NewDocument(index, 1.0)
	doc.Set("status", pinStatus.Status).
		Set("created", created).
		Set("cid", pinStatus.Pin.Cid).
		Set("name", escapeString(pinStatus.Pin.Name)).
		Set("origins", strings.Join(pinStatus.Pin.Origins, ",")).
		Set("meta", pinStatus.Pin.Meta.String()).
		Set("delegates", strings.Join(pinStatus.Delegates, ",")).
		Set("info", pinStatus.Info.String())
	return d.Index(doc)
}

func (d *pins) Patch(id string, fields map[string]interface{}) error {
	index := fmt.Sprintf("%s%s", PinsKeyPrefix, id)
	doc, err := d.Client.Get(index)
	if err != nil {
		return err
	}
	if doc == nil {
		return ErrNotFound
	}
	for key, val := range fields {
		var i interface{}
		switch key {
		case "status":
			i = val
		case "name":
			if name, ok := val.(string); ok {
				i = escapeString(name)
			}
		case "origins":
			if origins, ok := val.([]string); ok && origins != nil {
				i = strings.Join(origins, ",")
			}
		case "meta":
			if info, ok := val.(models.Info); ok && info != nil {
				i = info
			}
		case "#":
			i = val
		}
		if i != nil {
			doc.Set(key, i)
		}
	}
	return d.IndexOptions(redisearch.IndexingOptions{
		Partial: true,
		Replace: true,
	}, *doc)
}

func (d *pins) Get(id string) (models.PinStatus, error) {
	index := fmt.Sprintf("%s%s", PinsKeyPrefix, id)
	doc, err := d.Client.Get(index)
	if err != nil {
		return models.PinStatus{}, err
	}
	if doc == nil {
		return models.PinStatus{}, ErrNotFound
	}
	return parsePinStatusDoc(*doc), nil
}

type SortingKey struct {
	Field     string
	Ascending bool
}

func (d *pins) Find(
	cids, statuses []string,
	name string,
	before, after int64,
	match string,
	limit int, offset int,
	sortBy *SortingKey,
	meta models.Info,
) (models.Pins, error) {
	raw := []string{}
	if len(cids) > 0 {
		// filter by list of cids
		raw = append(
			raw,
			fmt.Sprintf("@cid:(%s) ", strings.Join(cids, "|")),
		)
	}
	if len(statuses) > 0 {
		// filter by list of statuses
		raw = append(
			raw,
			fmt.Sprintf(
				"@status:(%s)",
				strings.Join(statuses, "|"),
			),
		)
	}
	if name != "" {
		name = escapeString(name)
		switch match {
		case models.TextMatchExact.String():
			fallthrough
		case models.TextMatchIExact.String():
			name = fmt.Sprintf("(%s)", name)
		case models.TextMatchPartial.String():
			fallthrough
		case models.TextMatchIPartial.String():
			name = fmt.Sprintf("~(%s)", name)
		}
		// prefix search for name
		raw = append(raw, fmt.Sprintf("@name:%s", name))
	}
	if after > 0 && before > 0 {
		// filter within date range
		raw = append(
			raw,
			fmt.Sprintf("@created:[%d %d]", after, before),
		)
	} else if after > 0 {
		// filter by lower range
		raw = append(raw, fmt.Sprintf("@created:[(%d +inf]", after))
	} else if before > 0 {
		// filter by upper range
		raw = append(raw, fmt.Sprintf("@created:[-inf (%d]", before))
	}
	if len(meta) > 0 {
		var ss []string
		keys := meta.Keys()
		for _, k := range keys {
			ss = append(ss, escapeString(fmt.Sprintf("%s:%s", k, meta[k])))
		}
		raw = append(raw, fmt.Sprintf("@meta:{%s}",
			strings.Join(ss, "|")))
	}
	queryString := "*"
	if len(raw) > 0 {
		queryString = strings.Join(raw, " ")
	}
	q := redisearch.NewQuery(queryString)
	if limit > 0 {
		q.Limit(offset, limit)
	}
	if sortBy != nil {
		q.SetSortBy(sortBy.Field, sortBy.Ascending)
	}
	docs, total, err := d.Client.Search(q)
	if err != nil {
		return models.Pins{}, err
	}
	pins := []models.PinStatus{}
	for _, doc := range docs {
		pins = append(pins, parsePinStatusDoc(doc))
	}
	return models.Pins{
		Count:   len(pins),
		Total:   total,
		Results: pins,
	}, nil
}

func (d *pins) Delete(id string) error {
	index := fmt.Sprintf("%s%s", PinsKeyPrefix, id)
	return d.Client.DeleteDocument(index)
}

func parsePinStatusDoc(doc redisearch.Document) models.PinStatus {
	status := models.StatusUndefined
	if s, ok := doc.Properties["status"]; ok && s != nil {
		status, _ = models.ToStatus(s.(string))
	}
	created := ""
	if c, ok := doc.Properties["created"]; ok && c != nil {
		created = c.(string)
	}
	cid := ""
	if c, ok := doc.Properties["cid"]; ok && c != nil {
		cid = c.(string)
	}
	name := ""
	if n, ok := doc.Properties["name"]; ok && n != nil {
		name = n.(string)
		name = unescapeString(name)
	}
	origins := []string{}
	if o, ok := doc.Properties["origins"]; ok && o != nil {
		if s := o.(string); s != "" {
			origins = strings.Split(s, ",")
		}
	}
	meta := ""
	if m, ok := doc.Properties["meta"]; ok && m != nil {
		meta = m.(string)
	}
	delegates := []string{}
	if d, ok := doc.Properties["delegates"]; ok && d != nil {
		if s := d.(string); s != "" {
			delegates = strings.Split(s, ",")
		}
	}
	info := ""
	if i, ok := doc.Properties["info"]; ok && i != nil {
		info = i.(string)
	}
	return models.PinStatus{
		Id:      strings.TrimPrefix(doc.Id, PinsKeyPrefix),
		Status:  status.String(),
		Created: created,
		Pin: models.Pin{
			Cid:     cid,
			Name:    name,
			Origins: origins,
			Meta:    models.InfoFromString(meta),
		},
		Delegates: delegates,
		Info:      models.InfoFromString(info),
	}
}

func escapeString(s string) string {
	return TokenDelimitersRe.ReplaceAllString(s, `\$2`)
}

func unescapeString(s string) string {
	return strings.ReplaceAll(s, "\\", "")
}

func containsIndex(cli *redisearch.Client, idx string) (bool, error) {
	indexes, err := cli.List()
	if err != nil {
		return false, err
	}
	for _, index := range indexes {
		if index == idx {
			return true, nil
		}
	}
	return false, nil
}
