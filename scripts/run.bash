#!/bin/bash

RUN=$(command -v pins)

log()
{
    echo "$(date +"%F %T"): $*"
}

join_by()
{
    local d=$1; shift;
    local f=$1; shift;
    printf %s "$f" "${@/#/$d}"
}

usage()
{
    echo -e "$(basename "$0") [-h] [-c <configuration>] [-d <database address>]
    [-i <ipfs address>] [-p <port>] [-g grant]...
    -- program to start an Axis service

where:
    -a  set authenticator address; default is 'http://127.0.0.1:8080/.well-known/jwks.json'
    -c  configuration file location; default is '${HOME}/.axis/axis.toml'
    -d  set database address; default is '127.0.0.1:6379'
    -h  show this help text
    -i  set IPFS cluster multi-address; default is '/ip4/127.0.0.1/tcp/9094'
    -p  set listening port of HTTP API; default is 7070
    -g  add an authentication grant"
    exit
}

# START #

conf="${HOME}/.axis/axis.toml"
port=7070
db_addr="127.0.0.1:6379"
auth_addr="http://127.0.0.1:8080/.well-known/jwks.json"
ipfs_addr="/ip4/127.0.0.1/tcp/9094"
auth_grants=()

while getopts "ha:c:d:p:i:g:" opt; do
    case "$opt" in
    [h?]) usage
        ;;
    a) auth_addr="${OPTARG}"
        ;;
    c) conf="${OPTARG}"
        ;;
    d) db_addr="${OPTARG}"
        ;;
    i) ipfs_addr="${OPTARG}"
        ;;
    p) port="${OPTARG}"
        ;;
    g) auth_grants+=($OPTARG)
    esac
done

conf_dir=$(dirname ${conf})
if [ ! -d ${conf_dir} ]; then
    log $(mkdir -vp ${conf_dir})
fi

auth_grants_str="[]"
if [ ${#auth_grants[@]} -gt 0 ]; then
    auth_grants_str="[\"$(join_by '", "' ${auth_grants[@]})\"]"
fi

cat <<EOF > ${conf}
host="0.0.0.0"
port=${port}
read_timeout=30
write_timeout=30

database_addr="${db_addr}"
authenticator_addr="${auth_addr}"
authentication_grants=${auth_grants_str}

ipfs_cluster_ssl=false
ipfs_cluster_no_verify_cert=false
ipfs_cluster_api_addr="${ipfs_addr}"
ipfs_cluster_timeout=5
EOF
log "created '${conf}':
$(cat ${conf} | sed 's/^/\t/')"

$RUN --config-file="${conf}"
