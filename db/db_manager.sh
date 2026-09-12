#!/bin/bash

#########################################################################
# Script Name:   db_manager.sh
# Description:   Manage PostgreSQL service and database setup.
# Author:        Denis Pylypenko <den.pylypen@protonmail.com>
# Contributors:  None
# Created:       2026-07-28
# Last modified: 2026-08-30
# Version:       1.1.1
# License:       MIT License
# Repository:    https://github.com/denis1836/utils/tree/main/db_manager
#########################################################################

# navigate to script dir
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR" || exit 1

# default colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
ORANGE='\e[38;5;202m'
BLUE='\033[0;34m'
NC='\033[0m'

# color types
HIGHLIGHT=$GREEN
WARNING=$ORANGE
NOTICE=$YELLOW
ERROR=$RED

# configuration file
CONF_FILE="./db_manager.conf"

YES=false

# filtering messages based on verbosity var
msg() {
    local level="$1"; shift
    case "$VERBOSITY" in
        full)    echo -e "$@" ;;
        minimal) [[ "$level" == "minimal" || "$level" == "quiet" ]] && echo -e "$@" ;;
        quiet)   [[ "$level" == "quiet" ]] && echo -e "$@" ;;
    esac
}

affirm(){
    comm=${1:-"Are you sure?"}

    if [[ ${YES} == true ]]; then
        return 0
    fi

    while true; do 
        echo -e -n "${NC}${comm}?[y/n]${NC}" 
        read -r -n 1 ans
        echo -e -n "${NC}${NC}"
        case $ans in 
            [Yy] ) return 0;; 
            [Nn] ) return 1;; 
            * ) echo -e "${NC}\nInvalid argument provided.${NC}";; 
        esac
    done
}

# configuration file check
if [[ ! -f "$CONF_FILE" ]]; then
    echo -e "${ERROR}[ERROR]${NC} Configuration file is missing"
    if affirm "Do you want to create it?"; then
        cat <<EOF > "$CONF_FILE"
SERVICE_NAME=""
DB_NAME=""
DB_USER=""
DB_ADMIN_GROUP="postgres"
DB_HOST="localhost"
DB_PORT="5432"

SQL_DIR="./sql"
SQL_PATTERN='^[0-9][0-9]_.*\.sql$'

REQUIRED_TABLES=(

)

EXTRA_SCHEMAS=()

LOG_DIR="./db_manager_logs"
LOG_FILE_NAME_PATTERN="%Y-%m-%d_%H:%M:%S_db_manager.log"

VERBOSITY="full"
HELLO="quiet"
BYE="quiet"

BANNER="no"
BANNER_TEXT=""
EOF
    echo "Complete the configuration file with your db config data and run the script again."
    exit 0
    else
        exit 1
    fi
fi

# shellcheck disable=SC1090
source "$CONF_FILE"

print_banner() {
    case "$BANNER" in
        yes)
            echo -e "${NC}${BANNER_TEXT}${NC}"
        ;;
        no) : ;;
    esac
}

print_hello() {
    case $HELLO in
        full)
            echo -e "${NC}--------------------------------------------------------------------------${NC}"
            echo -e "${NC}Welcome to the ${HIGHLIGHT}${SERVICE_NAME}${NC} database managment script${NC}"
            echo -e "${NC}--------------------------------------------------------------------------${NC}"
            ;;
        minimal)
            echo -e "${HIGHLIGHT}${SERVICE_NAME}${NC} db manager" 
            ;;
        quiet) : ;;
    esac
}

print_bye() {
    case $BYE in
        full)
            echo -e ""
            echo -e "${HIGHLIGHT}[SUCCESS]${NC} Database setup completed!${NC}"
            echo -e "${NC}-------------------------------------------------------------${NC}"
            echo -e "${NC}You can now connect to the database using:${NC}"
            echo -e "${HIGHLIGHT}sudo -u ${DB_USER} psql -d ${DB_NAME}${NC}"
            echo -e "${NC}-------------------------------------------------------------${NC}"
            echo -e "${NC}Have a great day!${NC}"
            ;;
        minimal)
            echo -e "${NC}Bye!${NC}"
            ;;
        quiet) : ;;
    esac
}

# error msg
trap 'echo -e "\n${NC}${RED}[ERROR]${NC}\n===========================================================\n Installation failed - an unexpected error occurred.\n Check systemctl for possible errors. \n===========================================================\n${NC}"; exit 1' ERR
set -Eeuo pipefail

if [[ ! -d "$LOG_DIR" ]]; then
    mkdir -p "$LOG_DIR"
fi

CURRENT_LOG_NAME=$(date +"$LOG_FILE_NAME_PATTERN")
LOG_FILE="${LOG_DIR}/${CURRENT_LOG_NAME}"

if [[ ! -f "$LOG_FILE" ]]; then
    touch "$LOG_FILE" 2>/dev/null
fi

if [[ ! -w "${LOG_FILE}" ]]; then
    echo -e "${YELLOW}[WARNING]${NC} Cannot write to log file at \"${LOG_FILE}\".${NC}"
    affirm "Do you want to proceed without logging?" || {
        echo -e "${NC}Script aborted by user.${NC}"
        exit 1
    }
    echo -e "${NC}Proceeding without logging.${NC}"
    LOG_FILE=""
fi

if [[ -n "$LOG_FILE" ]]; then
    exec > >(tee >(sed -r 's/\x1b\[[0-9;]*m//g' | awk '{ print strftime("[%Y-%m-%d %H:%M:%S]"), $0 }' >> "$LOG_FILE"))
else
    exec > >(sed -r 's/\x1b\[[0-9;]*m//g' | awk '{ print strftime("[%Y-%m-%d %H:%M:%S]"), $0 }')
fi

print_banner
print_hello

# check os
if [[ "$(uname)" != "Linux" || $OSTYPE != linux* ]]; then
    echo -e "${YELLOW}[WARNING]${NC} This script is designed to run on Linux systems.${NC}"
    echo -e "${YELLOW}Proceeding may lead to unexpected behavior.${NC}"
    affirm "Do you want to proceed" || {
        echo -e "${NC}Script aborted by user.${NC}"
        exit 1
    }
fi

# check for psql
if ! command -v psql > /dev/null 2>&1; then
    echo -e "${RED}[ERROR]${NC} PostgreSQL is not installed.${NC}"
    echo -e "${NC}Run: \"sudo apt install postgresql\" to install it.${NC}"
    exit 1
fi

ACTION=""
USER_ACTION=""
GRANT_LEVEL=""

DB_RUN_AS=""
PSQL_RUNNER=""
TARGET_USER=""
PROMPT_PASSWORD=false

QUERY_TEXT=""
QUERY_FILE=""
QUERY_SOURCE=""

RUN_FROM=""
RUN_TO=""
RUN_ALL=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --start) ACTION="start"; shift ;;
        --stop) ACTION="stop"; shift ;;
        --restart) ACTION="restart"; shift ;;
        --status) ACTION="status"; shift ;;
        --drop-existing-db) ACTION="drop_db"; shift ;;
        --help) ACTION="help"; shift ;;
        -y|--yes) YES=true; shift ;;
 
        --user)
            ACTION="user"
            USER_ACTION="${2:?--user requires: add, remove, grant, or passwd}"
            shift 2
        ;;
 
        -U)
            DB_RUN_AS="${2:?-U requires a username}"
            shift 2
        ;;
 
        -u)
            TARGET_USER="${2:?-u requires a username}"
            shift 2
        ;;
 
        -p) PROMPT_PASSWORD=true; shift ;;
 
        --level)
            GRANT_LEVEL="${2:?--level requires: read, readwrite, or all}"
            shift 2
        ;;
 
        query)
            ACTION="query"
            shift
            if [[ "${1:-}" == "-f" ]]; then
                QUERY_FILE="${2:?-f requires a file path}"
                shift 2
            elif [[ "${1:-}" == "-" || $# -eq 0 ]]; then
                QUERY_SOURCE="stdin"
                [[ "${1:-}" == "-" ]] && shift
            else
                QUERY_TEXT="$1"
                shift
            fi
        ;;
 
        --run) ACTION="run"; shift ;;
 
        --from)
            RUN_FROM="${2:?--from requires a number}"
            shift 2
        ;;
        --to)
            RUN_TO="${2:?--to requires a number}"
            shift 2
        ;;
 
        --all) RUN_ALL=true; shift ;;
 
        --) shift; break ;;
 
        -*)
            echo -e "Error: unknown option '$1'" >&2
            echo -e "Run with --help to see available options" >&2
            exit 1
        ;;
 
        *)
            echo "Error: unexpected argument '$1'" >&2
            exit 1
        ;;
    esac
done


ACTION="${ACTION:-setup}"

if [[ "$ACTION" =~ ^(start|stop|restart)$ ]] && (( UID != 0 )); then
    echo -e "${RED}[ERROR]${NC} Managing the PostgreSQL service requires root or sudo.${NC}"
    exit 1
fi

case "$ACTION" in
    user)
        case "$USER_ACTION" in
            add|remove|grant|passwd) ;;
            *)
                echo "Error: --user must be one of: add, remove, grant, passwd" >&2
                exit 1
            ;;
        esac

        [[ -z "$TARGET_USER" ]] && {
            echo "Error: --user $USER_ACTION requires -u <name>" >&2
            exit 1
        }

        if [[ "$USER_ACTION" == "add" || "$USER_ACTION" == "passwd" ]] && [[ "$PROMPT_PASSWORD" != true ]]; then
            echo "Error: --user $USER_ACTION requires -p (password prompt)" >&2
            exit 1
        fi

        if [[ -n "$GRANT_LEVEL" && "$USER_ACTION" != "grant" ]]; then
            echo "Error: --level only applies to --user grant" >&2
            exit 1
        fi

        if [[ "$USER_ACTION" == "grant" ]]; then
            case "$GRANT_LEVEL" in
                read|readwrite|all) ;;
                *)
                    echo "Error: --user grant requires --level read|readwrite|all" >&2
                    exit 1
                ;;
            esac
        fi
    ;;

    query)
        if [[ -z "$QUERY_TEXT" && -z "$QUERY_FILE" && "$QUERY_SOURCE" != "stdin" ]]; then
            echo "Error: query requires SQL text, -f <file>, or stdin input" >&2
            exit 1
        fi
        if [[ "$PROMPT_PASSWORD" == true && "$QUERY_SOURCE" == "stdin" ]]; then
            echo "Error: -p (password prompt) can't be combined with stdin query input" >&2
            exit 1
        fi
    ;;

    run)
        if [[ "$RUN_ALL" == true && ( -n "$RUN_FROM" || -n "$RUN_TO" ) ]]; then
            echo "Error: --all cannot be combined with --from/--to" >&2
            exit 1
        fi
        if [[ "$RUN_ALL" != true && -z "$RUN_FROM" && -z "$RUN_TO" ]]; then
            echo "Error: --run requires --all, or --from/--to" >&2
            exit 1
        fi
    ;;
esac


determine_psql_runner() {
    local target_user="${DB_RUN_AS:-$DB_USER}"
 
    if [[ "$(id -un)" == "${target_user}" ]]; then
        PSQL_RUNNER=""
        return 0
    fi
 
    if (( UID == 0 )); then
        PSQL_RUNNER="sudo -u ${target_user}"
        return 0
    fi
 
    if sudo -n -u "${target_user}" true 2>/dev/null; then
        PSQL_RUNNER="sudo -u ${target_user}"
        return 0
    fi
 
    if id -nG "$(id -un)" 2>/dev/null | tr ' ' '\n' | grep -qx "${DB_ADMIN_GROUP:-postgres}"; then
        if sudo -n -u "${target_user}" true 2>/dev/null; then
            PSQL_RUNNER="sudo -u ${target_user}"
            return 0
        fi
    fi
 
    echo -e "${ERROR}[ERROR]${NC} Cannot run as \"${target_user}\" without a password prompt.${NC}"
    echo -e "${NC}Options:${NC}"
    echo -e "${NC} - run this script as \"${target_user}\" directly${NC}"
    echo -e "${NC} - run with sudo${NC}"
    echo -e "${NC} - ask an admin to add a NOPASSWD sudo rule for user \"${target_user}\"${NC}"
    return 1
}

show_help() {
    echo -e "${NC}Usage: ./db_manager.sh [ACTION] [OPTIONS]${NC}"
    echo -e "${NC}${NC}"
    echo -e "${NC}Manage PostgreSQL service and database setup.${NC}"
    echo -e "${NC}${NC}"
    echo -e "${NC}Service actions:${NC}"
    echo -e "${HIGHLIGHT}  --start${NC}                     Start PostgreSQL service (needs root)"
    echo -e "${HIGHLIGHT}  --stop${NC}                      Stop PostgreSQL service (needs root)"
    echo -e "${HIGHLIGHT}  --restart${NC}                   Restart PostgreSQL service (needs root)"
    echo -e "${HIGHLIGHT}  --status${NC}                    Check PostgreSQL daemon and database state"
    echo -e "${NC}${NC}"
    echo -e "${NC}Database setup:${NC}"
    echo -e "${HIGHLIGHT}  (no action)${NC}                 Run the full SQL setup (all files in SQL_DIR)"
    echo -e "${HIGHLIGHT}  --drop-existing-db${NC}          Drop the existing database"
    echo -e "${HIGHLIGHT}  --run --all${NC}                 Execute all SQL files"
    echo -e "${HIGHLIGHT}  --run --from N [--to M]${NC}     Execute SQL files numbered N..M (or N..end)"
    echo -e "${HIGHLIGHT}  --run --to M${NC}                Execute SQL files numbered 0..M"
    echo -e "${NC}${NC}"
    echo -e "${NC}User management:${NC}"
    echo -e "${HIGHLIGHT}  --user add -u <name> -p${NC}     Create a role (prompts for password)"
    echo -e "${HIGHLIGHT}  --user remove -u <name>${NC}     Drop a role"
    echo -e "${HIGHLIGHT}  --user passwd -u <name>${NC}     Change a role's password (prompts twice)"
    echo -e "${HIGHLIGHT}  --user grant -u <name> --level <read|readwrite|all>${NC}"
    echo -e "${NC}                                 Grant privileges on ${DB_NAME} to a role${NC}"
    echo -e "${NC}${NC}"
    echo -e "${NC}Ad-hoc queries:${NC}"
    echo -e "${HIGHLIGHT}  query \"<SQL>\"${NC}               Run inline SQL"
    echo -e "${HIGHLIGHT}  query -f <file>${NC}             Run SQL from a file"
    echo -e "${HIGHLIGHT}  query -${NC} / ${HIGHLIGHT}query${NC}                Run SQL from stdin (for piping/automation)"
    echo -e "${NC}${NC}"
    echo -e "${NC}Common options:${NC}"
    echo -e "${HIGHLIGHT}  -U <system-user>${NC}            Run psql as this system user instead of DB_USER"
    echo -e "${HIGHLIGHT}  -y, --yes${NC}                   Assume 'yes' to all confirmation prompts"
    echo -e "${NC}${NC}"
    echo -e "${NC}Note: -p always prompts interactively and never accepts a value inline,${NC}"
    echo -e "${NC}      to avoid passwords leaking into shell history or 'ps aux'.${NC}"
    echo -e "${NC}      For unattended automation (cron, other scripts), use ~/.pgpass instead.${NC}"
    exit 0
}

run_status() {
    echo -e "${NOTICE}[NOTICE]${NC} Checking PostgreSQL daemon status...${NC}"
    systemctl is-active --quiet postgresql || {
        echo -e "${ERROR}[ERROR]${NC} PostgreSQL daemon is not running.${NC}"
        exit 1
    }
    echo -e "${HIGHLIGHT}[OK]${NC} PostgreSQL daemon is active.${NC}"

    echo -e "${NOTICE}[NOTICE]${NC} Checking ${DB_NAME} database state...${NC}"
    
    timeout 10 pg_isready -d "${DB_NAME}" -q || {
        echo -e "${ERROR}[ERROR]${NC} Database '${DB_NAME}' is not responding or timed out.${NC}"
        exit 1
    }
    echo -e "${HIGHLIGHT}[OK]${NC} Database '${DB_NAME}' is ready.${NC}"
    exit 0
}

run_start() {
    echo -e "${NOTICE}[NOTICE]${NC} Starting PostgreSQL...${NC}"
    timeout 10 systemctl start postgresql || {
        echo -e "${ERROR}[ERROR]${NC} PostgreSQL start timed out.${NC}"
        exit 1
    }
    if [[ "$(systemctl is-active postgresql)" != "active" ]]; then
        echo -e "${ERROR}[ERROR]${NC} PostgreSQL failed to start. ${NC}"
        exit 1
    fi
    echo -e "${HIGHLIGHT}[OK]${NC} PostgreSQL started successfully.${NC}"
    exit 0
}

run_stop() {
    echo -e "${NOTICE}[NOTICE]${NC} Stopping PostgreSQL...${NC}"
    timeout 10 systemctl stop postgresql || {
        echo -e "${ERROR}[ERROR]${NC} PostgreSQL stop timed out.${NC}"
        exit 1
    }
    if [[ "$(systemctl is-active postgresql)" == "active" ]]; then
        echo -e "${ERROR}[ERROR]${NC} PostgreSQL failed to stop. ${NC}"
        exit 1
    fi
    echo -e "${HIGHLIGHT}[OK]${NC} PostgreSQL stopped successfully.${NC}"
    exit 0
}

run_restart() {
    echo -e "${NOTICE}[NOTICE]${NC} Restarting PostgreSQL...${NC}"
    timeout 10 systemctl restart postgresql || {
        echo -e "${ERROR}[ERROR]${NC} PostgreSQL restart timed out.${NC}"
        exit 1
    }
    if [[ "$(systemctl is-active postgresql)" != "active" ]]; then
        echo -e "${ERROR}[ERROR]${NC} PostgreSQL failed to restart. ${NC}"
        exit 1
    fi
    echo -e "${HIGHLIGHT}[OK]${NC} PostgreSQL restarted successfully.${NC}"
    exit 0
}

check_psql_daemon() {
    if [[ "$(systemctl is-active postgresql)" != "active" ]]; then
        echo -e "${NOTICE}[NOTICE]${NC} PostgreSQL is not running. Starting...${NC}"
 
        timeout 10 systemctl start postgresql || {
            echo -e "${ERROR}[ERROR]${NC} PostgreSQL start timed out.${NC}"
            exit 1
        }
 
        if [[ "$(systemctl is-active postgresql)" != "active" ]]; then
            echo -e "${ERROR}[ERROR]${NC} PostgreSQL failed to start. ${NC}"
            exit 1
        fi
    fi
}

run_user_action() {
case "$USER_ACTION" in
    add)
        local user_login="${TARGET_USER}"
        local user_pass=""
        
        if [[ $PROMPT_PASSWORD ]]; then
            while true; do 
                user_entry1=""
                user_entry2=""

                read -r -s -p "Enter user password: " user_entry1
                read -r -s -p "Confirm password: " user_entry2

                if [[ "$user_entry1" != "$user_entry2" ]]; then
                    echo -e "${ERROR}Error: ${NC}Passwords do not match.${NC}"
                    continue
                fi
                
                user_pass="$user_entry2"
                break
            done
        fi

        echo -e "${NC}Adding user \"${user_login}\"...${NC}"
    
        ${PSQL_RUNNER} psql -v ON_ERROR_STOP=1 -c "
            DO \$$
            BEGIN
            IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '$user_login') THEN
                CREATE ROLE '$user_login' WITH LOGIN PASSWORD '$user_pass';
                ALTER ROLE '$user_login' WITH LOGIN;
            END IF;
            END
            \$$;
        "

        echo -e "${NC}Added user: \"${user_login}\".${NC}"
        echo -e "${NOTICE}[NOTICE]${NC}Remeber to keep your password private."
    ;;

    remove)
        local user_login="${TARGET_USER}"
        local user_pass=""
        
        if [[ $PROMPT_PASSWORD ]]; then
            while true; do 
                user_entry1=""
                user_entry2=""

                read -r -s -p "Enter user password: " user_entry1
                read -r -s -p "Confirm password: " user_entry2

                if [[ "$user_entry1" != "$user_entry2" ]]; then
                    echo -e "${ERROR}Error: ${NC}Passwords do not match.${NC}"
                    continue
                fi
                
                user_pass="$user_entry2"
                break
            done
        fi

        echo -e "${NC}Removing user \"${user_login}\"...${NC}"
    
        ${PSQL_RUNNER} psql -v ON_ERROR_STOP=1 -c "
            DO \$$
            BEGIN
            IF EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '$user_login') THEN
                DROP ROLE '$user_login';
            END IF;
            END
            \$$;
        "

        echo -e "${NC}Removed user: \"${user_login}\".${NC}"
    ;;

    grant)
        case "${GRANT_LEVEL}" in 
            read)
                affirm "Do you want to grant read privileges on the \"${DB_NAME}\" database to user \"${TARGET_USER}\"" && {
                    ${PSQL_RUNNER} psql -v ON_ERROR_STOP=1 -c "
                        GRANT SELECT ON DATABASE ${DB_NAME} TO ${TARGET_USER};
                    "
                    echo -e "${HIGHLIGHT}[OK]${NC}Privileges granted to user \"${TARGET_USER}\".${NC}"
                }
            ;;
    
            readwrite)
                affirm "Do you want to grant read privileges on the \"${DB_NAME}\" database to user \"${TARGET_USER}\"" && {
                    ${PSQL_RUNNER} psql -v ON_ERROR_STOP=1 -c "
                        GRANT SELECT INSERT UPDATE DELETE REFENCES ON DATABASE ${DB_NAME} TO ${TARGET_USER};
                    "
                    echo -e "${HIGHLIGHT}[OK]${NC}Privileges granted to user \"${TARGET_USER}\".${NC}"
                }

            ;;

            all)
                affirm "Do you want to grant all privileges on the \"${DB_NAME}\" database to user \"${TARGET_USER}\"" && {
                    ${PSQL_RUNNER} psql -v ON_ERROR_STOP=1 -c "
                        GRANT ALL PRIVILEGES ON DATABASE ${DB_NAME} TO ${TARGET_USER};
                        GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO ${TARGET_USER};
                        GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO ${TARGET_USER};
                        GRANT USAGE ON SCHEMA cron TO ${TARGET_USER};
                    "
                    echo -e "${HIGHLIGHT}[OK]${NC}Privileges granted to user \"${TARGET_USER}\".${NC}"
                }
            ;;
            *)
                echo "Error: Invalid --level value provided" >&2
            ;;
        esac
    ;;

    passwd)
        local user_login="${TARGET_USER}"
        local curr_user_pass=""
        local new_user_pass=""

        read -r -s -p "Enter current role password: " curr_user_pass
        echo ""

        if ! PGPASSWORD="${curr_user_pass}" ${PSQL_RUNNER} psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${user_login}" -d "${DB_NAME}" -v ON_ERROR_STOP=1 -c "SELECT 1;" >/dev/null 2>&1; then
            echo -e "${ERROR}[ERROR]${NC} Current password does not match or user cannot login to ${DB_NAME} on ${DB_HOST}.${NC}"
            exit 1 
        fi

        read -r -s -p "Enter new role password: " new_user_pass       
        echo ""

        echo -e "${NC}Changing \"${user_login}\" user password...${NC}"
    
        ${PSQL_RUNNER} psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" -v ON_ERROR_STOP=1 -c "
            DO \$$
            BEGIN
            IF EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '${user_login}') THEN
                EXECUTE format('ALTER ROLE %I WITH PASSWORD %L', '${user_login}', '${new_user_pass}');
            END IF;
            END
            \$$;
        "

        echo -e "${NC}User password was changed${NC}"
        echo -e "${NOTICE}[NOTICE]${NC}Remeber to keep your password private."
    ;;
esac
}

run_query() {
    if [[ -n "$QUERY_FILE" ]]; then
        ${PSQL_RUNNER} psql -v ON_ERROR_STOP=1 -d "${DB_NAME}" -f "${QUERY_FILE}"
    elif [[ "$QUERY_SOURCE" == "stdin" ]]; then
        ${PSQL_RUNNER} psql -v ON_ERROR_STOP=1 -d "${DB_NAME}" -f -
    else
        ${PSQL_RUNNER} psql -v ON_ERROR_STOP=1 -d "${DB_NAME}" -c "${QUERY_TEXT}"
    fi
}

run_drop_db() {
    if ${PSQL_RUNNER} psql -v ON_ERROR_STOP=1 -tAc "SELECT 1 FROM pg_database WHERE datname='${DB_NAME}';" | grep -q 1; then
        affirm "Do you want to drop existing database \"${DB_NAME}\"? This action will destroy ALL the data." || {
            echo -e "${NC}Script aborted by user.${NC}"
            exit 1
        }
        
        echo -e "\n${NOTICE}Terminating active connections to \"${DB_NAME}\"...${NC}"
        ${PSQL_RUNNER} psql -v ON_ERROR_STOP=1 -c "
            SELECT pg_terminate_backend(pid) 
            FROM pg_stat_activity 
            WHERE datname = '${DB_NAME}' AND pid <> pg_backend_pid();
        "

        echo -e "${WARNING}Dropping existing database \"${DB_NAME}\"...${NC}"
        ${PSQL_RUNNER} psql -v ON_ERROR_STOP=1 -c "DROP DATABASE IF EXISTS ${DB_NAME};"
        echo -e "${HIGHLIGHT}Existing database dropped.${NC}"
        exit 0
    else
        echo -e "${HIGHLIGHT}Database \"${DB_NAME}\" does not exist. Nothing to drop.${NC}"
        exit 0
    fi
}

verify_db_creation(){
    if ! ${PSQL_RUNNER} psql -v ON_ERROR_STOP=1 -tAc \
      "SELECT 1 FROM pg_database WHERE datname='${DB_NAME}';" | grep -q 1; then
        echo -e "${RED}[ERROR]${NC} Database ${DB_NAME} was ${RED}NOT${NC} created successfully.${NC}"
        echo -e "${NC}Database setup incomplete. Check SQL files for possible errors or run the script again.${NC}"
        exit 1
    fi
}

verify_db_tables() {
    for table in "${REQUIRED_TABLES[@]}"; do
        if ! ${PSQL_RUNNER} psql -v ON_ERROR_STOP=1 -d "${DB_NAME}" -tAc \
            "SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND lower(table_name)=lower('${table}');" | grep -q 1; then
            echo -e "${ERROR}[ERROR]${NC} Missing table \"${table}\"${NC}"
            echo -e "${NC}Database setup incomplete. Check SQL files for possible errors and run the script again.${NC}"
            exit 1
        fi
    done
}

check_psql_conn() {
    if ! ${PSQL_RUNNER} psql -v ON_ERROR_STOP=1 -tAc "SELECT 1;" > /dev/null 2>&1; then
        echo -e "${RED}[ERROR]${NC} Cannot connect to PostgreSQL as user \"${DB_USER}\".${NC}"
        echo -e "${NC}Check PostgreSQL installation and user permissions.${NC}"
        exit 1
    fi
}

check_if_db_exists() {
    if ${PSQL_RUNNER} psql -v ON_ERROR_STOP=1 -tAc \
      "SELECT 1 FROM pg_database WHERE datname='${DB_NAME}';" | grep -q 1; then
        echo -e "${YELLOW}[WARNING]${NC} Database \"${DB_NAME}\" already exists.${NC}"
        affirm "Do you want to proceed and possibly overwrite existing database" || {
            echo -e "${NC}Script aborted by user.${NC}"
            exit 1
        }
    fi
}

sql_file_number() {
    [[ "$1" =~ ^([0-9]+)_ ]] && echo "${BASH_REMATCH[1]}"
}

discover_sql_files() {
    SQL_FILES=()
    local f base

    for f in "${SQL_DIR}"/*; do
        [[ -e "$f" ]] || continue
        [[ -f "$f" ]] || continue
        base="$(basename "$f")"

        if [[ "$base" =~ $SQL_PATTERN ]]; then
            SQL_FILES+=("$base")
        fi
    done

    if [[ ${#SQL_FILES[@]} -gt 0 ]]; then
        mapfile -t SQL_FILES < <(printf '%s\n' "${SQL_FILES[@]}" | sort)
    fi
}

filter_sql_files() {
    FILES_TO_RUN=()
    local f raw_num num

    for f in "${SQL_FILES[@]}"; do
        raw_num=$(sql_file_number "$f")
        [[ -z "$raw_num" ]] && continue

        num=$((10#$raw_num))

        if [[ "$RUN_ALL" != true ]]; then
            [[ -n "$RUN_FROM" && $num -lt $((10#$RUN_FROM)) ]] && continue
            [[ -n "$RUN_TO"   && $num -gt $((10#$RUN_TO))   ]] && continue
        fi

        FILES_TO_RUN+=("$f")
    done
}

execute_sql_files() {
    local list=("$@")
    msg full "${HIGHLIGHT}${#list[@]}${NC} SQL files will be executed.${NC}"

    local sql num
    for sql in "${list[@]}"; do
        num=$((10#$(sql_file_number "$sql")))
        msg full "${NOTICE}>[PSQL]${NC} Executing ${sql}...${NC}"

        if [[ $num -eq 0 ]]; then
            if ! sudo -u postgres psql -v ON_ERROR_STOP=1 -tAc \
              "SELECT 1 FROM pg_database WHERE datname='${DB_NAME}';" | grep -q 1; then
                ${PSQL_RUNNER} psql -v ON_ERROR_STOP=1 -f "${SQL_DIR}/${sql}"
            else
                msg full "${WARNING}[SKIP]${NC} Database already exists, skipping ${sql}.${NC}"
            fi
        else
            ${PSQL_RUNNER} psql -v ON_ERROR_STOP=1 -d "${DB_NAME}" -f "${SQL_DIR}/${sql}"
        fi
    done

    msg full "${HIGHLIGHT}[PSQL]${NC} All SQL files were executed.${NC}"
}

run_setup() {
    check_psql_conn
    check_if_db_exists

    discover_sql_files
    echo -e "${NC}Proceeding to create the database...${NC}"
    execute_sql_files "${SQL_FILES[@]}"

    echo -e "${NC}Verifying database integrity...${NC}"
    verify_db_creation
    verify_db_tables
}

if [[ "$ACTION" =~ ^(setup|user|query|run|drop_db)$ ]]; then
    if ! id -u "${DB_USER}" > /dev/null 2>&1; then
        echo -e "${ERROR}[ERROR]${NC} PostgreSQL user \"${DB_USER}\" does not exist.${NC}"
        echo -e "${NC}Create the user with: \"sudo -u postgres createuser ${DB_USER}\"${NC}"
        exit 1
    fi

    check_psql_daemon
    determine_psql_runner || exit 1
fi

case "$ACTION" in
    help) show_help ;;
    status) run_status ;;
    start) run_start ;;
    stop) run_stop ;;
    restart) run_restart ;;
    user) run_user_action ;;
    query) run_query ;;
    run)
        discover_sql_files
        filter_sql_files
        execute_sql_files "${FILES_TO_RUN[@]}"
    ;;
    drop_db) run_drop_db ;;
    setup) run_setup ;;
esac

print_bye
echo ""
exit 0
