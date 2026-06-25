#!/usr/bin/env bash
# =============================================================================
# ezbookkeeping 生产部署脚本
#
# 用法:
#   ./deploy.sh              完整部署（拉代码 → 构建 → 重启）
#   ./deploy.sh start        启动服务
#   ./deploy.sh stop         停止服务
#   ./deploy.sh restart      重启服务
#   ./deploy.sh status       查看服务状态
#   ./deploy.sh logs         查看服务实时日志
#   ./deploy.sh build        仅构建（不重启）
#   ./deploy.sh check        仅环境检查
#   ./deploy.sh rollback     回滚到上一个版本
#
# 参数:
#   --skip-frontend          跳过前端构建
#   --skip-backend           跳过后端构建
#
# 最低服务器要求: 1 核 CPU / 1 GB RAM
# =============================================================================

# 防止被 source 执行（而非直接执行）
if [[ "${BASH_SOURCE[0]}" != "${0}" ]]; then
    echo "错误: 此脚本不应被 source，请直接执行: ./deploy.sh"
    return 1 2>/dev/null || exit 1
fi

set -o pipefail

# =============================================================================
# 配置变量
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
APP_NAME="ezbookkeeping"
BINARY_NAME="ezbookkeeping"
REPO_CONFIG="conf/ezbookkeeping.ini"
USER_CONFIG="conf/ezbookkeeping.user.ini"
PID_FILE=".deploy.pid"
DEPLOY_LOG="deploy.log"
OLD_BINARY="${BINARY_NAME}.old"
NEW_BINARY="${BINARY_NAME}.new"
DIST_DIR="dist"
PUBLIC_DIR="public"
MAX_DEPLOY_LOGS=10

# 默认值（会被配置文件覆盖）
DEFAULT_PORT=8080
HEALTH_CHECK_TIMEOUT=30
HEALTH_CHECK_INTERVAL=2
GRACEFUL_TIMEOUT=15

# systemd
SYSTEMD_SERVICE="ezbookkeeping.service"
SYSTEMD_SERVICE_PATH="/etc/systemd/system/${SYSTEMD_SERVICE}"

# Nginx
NGINX_CONF_NAME="ezbookkeeping"
NGINX_SITES_AVAILABLE="/etc/nginx/sites-available/${NGINX_CONF_NAME}"
NGINX_SITES_ENABLED="/etc/nginx/sites-enabled/${NGINX_CONF_NAME}"
NGINX_CONF_NGINX="/etc/nginx/conf.d/${NGINX_CONF_NAME}.conf"  # RHEL 系

# 选项
SKIP_FRONTEND=false
SKIP_BACKEND=false

cd "$SCRIPT_DIR"

# =============================================================================
# systemd 检测
# =============================================================================

# 检查系统是否支持 systemd
has_systemd() {
    [[ -d /run/systemd/system ]]
}

# 检查 systemd 服务文件是否已安装
is_systemd_installed() {
    [[ -f "$SYSTEMD_SERVICE_PATH" ]]
}

# 检查 systemd 服务是否已启用
is_systemd_enabled() {
    systemctl is-enabled "$SYSTEMD_SERVICE" &>/dev/null
}

# 获取 systemd 服务状态
get_systemd_status() {
    systemctl is-active "$SYSTEMD_SERVICE" 2>/dev/null || echo "unknown"
}

# 检测 Nginx 安装方式（sites-available vs conf.d）
detect_nginx_config_dir() {
    if [[ -d "/etc/nginx/sites-available" ]]; then
        echo "sites"
    elif [[ -d "/etc/nginx/conf.d" ]]; then
        echo "confd"
    else
        echo "none"
    fi
}

# =============================================================================
# 颜色与日志
# =============================================================================

# 颜色码
COLOR_RESET='\033[0m'
COLOR_RED='\033[31m'
COLOR_GREEN='\033[32m'
COLOR_YELLOW='\033[33m'
COLOR_BLUE='\033[34m'
COLOR_CYAN='\033[36m'
COLOR_BOLD='\033[1m'

# 检测是否支持颜色输出
if [[ -t 1 ]] && [[ -n "$TERM" ]] && [[ "$TERM" != "dumb" ]]; then
    USE_COLOR=true
else
    USE_COLOR=false
fi

now() {
    date '+%Y-%m-%d %H:%M:%S'
}

_log() {
    local level="$1"
    local color="$2"
    local message="$3"
    local timestamp
    timestamp="$(now)"

    # 写入部署日志文件
    echo "[$timestamp] [$level] $message" >> "$DEPLOY_LOG"

    # 输出到终端（带颜色）
    if $USE_COLOR; then
        printf "${color}[%s] [%s]${COLOR_RESET} %s\n" "$timestamp" "$level" "$message"
    else
        echo "[$timestamp] [$level] $message"
    fi
}

log_info()    { _log "INFO"    "$COLOR_BLUE"   "$1"; }
log_success() { _log "SUCCESS" "$COLOR_GREEN"  "$1"; }
log_warn()    { _log "WARN"    "$COLOR_YELLOW" "$1"; }
log_error()   { _log "ERROR"   "$COLOR_RED"    "$1"; }
log_step()    { _log "STEP"    "$COLOR_CYAN${COLOR_BOLD}" "$1"; }

# 打印分隔线
print_separator() {
    if $USE_COLOR; then
        printf "${COLOR_CYAN}%*s${COLOR_RESET}\n" 60 "" | tr ' ' '─'
    else
        printf '%*s\n' 60 "" | tr ' ' '-'
    fi
}

# =============================================================================
# 工具函数
# =============================================================================

# 获取配置值（从用户配置或默认配置中读取）
get_config_value() {
    local section="$1"
    local key="$2"
    local config_file="${3:-$USER_CONFIG}"

    # 先尝试用户配置，不存在则用仓库默认配置
    if [[ ! -f "$config_file" ]]; then
        config_file="$REPO_CONFIG"
    fi

    if [[ ! -f "$config_file" ]]; then
        echo ""
        return
    fi

    # 简单的 INI 解析：在指定 section 中查找 key
    awk -F '=' -v sec="$section" -v k="$key" '
        /^\[.*\]/ { in_section = ($0 == "[" sec "]") }
        in_section && $1 ~ /^[[:space:]]*'"$key"'[[:space:]]*$/ {
            val = $2
            gsub(/^[[:space:]]+|[[:space:]]+$/, "", val)
            print val
            exit
        }
    ' "$config_file"
}

# 获取服务端口
get_server_port() {
    local port
    port="$(get_config_value "server" "http_port")"
    echo "${port:-$DEFAULT_PORT}"
}

# 获取进程 PID（优先用 PID 文件，其次用 pgrep）
get_pid() {
    if [[ -f "$PID_FILE" ]]; then
        local pid
        pid="$(cat "$PID_FILE" 2>/dev/null)"
        if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
            echo "$pid"
            return 0
        fi
    fi
    # 回退：通过进程名查找
    pgrep -f "$BINARY_NAME server run" 2>/dev/null | head -1
}

# 检查服务是否在运行
is_running() {
    local pid
    pid="$(get_pid)"
    [[ -n "$pid" ]]
}

# 检查命令是否存在
check_command() {
    local cmd="$1"
    local name="${2:-$cmd}"
    local min_version="$3"

    if ! command -v "$cmd" &>/dev/null; then
        log_error "未找到 $name，请先安装"
        return 1
    fi

    if [[ -n "$min_version" ]]; then
        local version
        case "$cmd" in
            go)
                version="$(go version 2>/dev/null | sed -n 's/.*go\([0-9]\+\.[0-9]\+\).*/\1/p' | head -1)"
                ;;
            node)
                version="$(node -v 2>/dev/null | sed -n 's/v\([0-9]\+\).*/\1/p' | head -1)"
                ;;
            *)
                version=""
                ;;
        esac
        if [[ -n "$version" ]] && [[ "$(echo "$version $min_version" | awk '{print ($1 < $2)}')" -eq 1 ]]; then
            log_warn "$name 版本 $version < 建议版本 $min_version，可能存在问题"
        fi
    fi
    return 0
}

# 轮换部署日志（保留最近 N 次）
rotate_deploy_log() {
    if [[ -f "$DEPLOY_LOG" ]]; then
        local lines
        lines="$(wc -l < "$DEPLOY_LOG")"
        if [[ "$lines" -gt 5000 ]]; then
            tail -n 2000 "$DEPLOY_LOG" > "${DEPLOY_LOG}.tmp"
            mv "${DEPLOY_LOG}.tmp" "$DEPLOY_LOG"
        fi
    fi
}

# =============================================================================
# 环境检查
# =============================================================================

do_check() {
    log_step "环境检查"

    local errors=0

    log_info "检查必要依赖..."

    check_command "git"   "Git"    || ((errors++))
    check_command "go"    "Go"     "1.26" || ((errors++))
    check_command "node"  "Node.js" "20" || ((errors++))
    check_command "npm"   "npm"    || ((errors++))

    # C 编译器检查（CGO 需要真正的 C 编译器，而非同名 npm 工具）
    local cc_path
    cc_path="$(which cc 2>/dev/null || true)"
    if [[ -n "$cc_path" ]]; then
        # 验证 cc 是否是真正的 C 编译器
        if echo 'int main(){return 0;}' | "$cc_path" -x c - -o /dev/null 2>/dev/null; then
            log_info "  C 编译器: $cc_path ✓"
        else
            log_warn "  \$ cc 指向非 C 编译器 ($cc_path)，尝试查找系统编译器..."
            # 回退策略
            if [[ -x "/usr/bin/cc" ]] && /usr/bin/cc --version &>/dev/null; then
                export CC="/usr/bin/cc"
                log_info "  已设置 CC=/usr/bin/cc"
            elif command -v gcc &>/dev/null; then
                export CC="$(which gcc)"
                log_info "  已设置 CC=$CC"
            else
                log_error "  未找到可用的 C 编译器，CGO 构建将失败"
                ((errors++))
            fi
        fi
    else
        log_warn "  未找到 cc 命令，CGO 构建可能失败"
    fi

    # 显示版本信息
    if command -v go &>/dev/null; then
        log_info "  Go:      $(go version | head -1)"
    fi
    if command -v node &>/dev/null; then
        log_info "  Node.js: $(node -v)"
    fi
    if command -v npm &>/dev/null; then
        log_info "  npm:     $(npm -v)"
    fi
    if command -v gcc &>/dev/null; then
        log_info "  GCC:     $(gcc --version | head -1)"
    fi
    if command -v git &>/dev/null; then
        log_info "  Git:     $(git --version | head -1)"
    fi

    # 检查内存
    if [[ -f /proc/meminfo ]]; then
        local mem_total_kb
        mem_total_kb="$(grep MemTotal /proc/meminfo | awk '{print $2}')"
        local mem_total_mb=$((mem_total_kb / 1024))
        log_info "  总内存:  ${mem_total_mb}MB"
        if [[ "$mem_total_mb" -lt 512 ]]; then
            log_warn "  内存不足 512MB，构建过程可能较慢或失败"
        fi
    fi

    # 检查磁盘空间（兼容 Linux 和 macOS）
    local available_space
    if df -BG . &>/dev/null 2>&1; then
        # Linux
        available_space="$(df -BG . | tail -1 | awk '{print $4}' | tr -d 'G')"
    else
        # macOS (df -g gives 1024MB blocks)
        available_space="$(df -g . | tail -1 | awk '{print $4}')"
        # 转换为 GB（大致）
        available_space="$((available_space * 1024 / 1000))"
    fi
    if [[ -n "$available_space" ]] && [[ "$available_space" =~ ^[0-9]+$ ]]; then
        log_info "  可用磁盘: ${available_space}GB"
        if [[ "$available_space" -lt 2 ]]; then
            log_warn "  可用磁盘空间不足 2GB"
        fi
    fi

    if [[ "$errors" -gt 0 ]]; then
        log_error "环境检查未通过（${errors} 个错误），请安装缺失的依赖"
        return 1
    fi

    log_success "环境检查通过"
    return 0
}

# =============================================================================
# 代码更新
# =============================================================================

do_pull() {
    log_step "代码更新"

    if [[ ! -d ".git" ]]; then
        log_warn "未检测到 .git 目录，跳过 git pull"
        log_info "请确保已将完整仓库上传到服务器"
        return 0
    fi

    # 记录拉取前的 commit
    local before_commit
    before_commit="$(git rev-parse --short HEAD 2>/dev/null)"
    log_info "当前 commit: ${before_commit:-未知}"

    # 暂存本地修改（防止 git pull 冲突）
    if ! git diff --quiet 2>/dev/null; then
        log_warn "检测到本地修改，执行 git stash"
        git stash save "auto-stash before deploy $(now)" 2>/dev/null || true
    fi

    log_info "执行 git pull..."
    if ! git pull --ff-only 2>&1; then
        log_error "git pull 失败，请手动处理冲突"
        return 1
    fi

    local after_commit
    after_commit="$(git rev-parse --short HEAD 2>/dev/null)"

    if [[ "$before_commit" == "$after_commit" ]]; then
        log_info "代码无更新（commit: $after_commit）"
    else
        log_success "代码已更新: ${before_commit:-无} → $after_commit"

        # 显示变更摘要
        if [[ -n "$before_commit" ]]; then
            log_info "变更文件:"
            git diff --stat "$before_commit..$after_commit" 2>/dev/null | head -20 | while read -r line; do
                log_info "  $line"
            done
        fi
    fi

    return 0
}

# =============================================================================
# 配置管理
# =============================================================================

do_config_init() {
    log_step "配置检查"

    # 确保仓库默认配置存在
    if [[ ! -f "$REPO_CONFIG" ]]; then
        log_error "仓库默认配置 $REPO_CONFIG 不存在"
        return 1
    fi

    # 首次部署：复制用户配置
    if [[ ! -f "$USER_CONFIG" ]]; then
        log_info "首次部署，创建用户配置文件..."
        cp "$REPO_CONFIG" "$USER_CONFIG"
        log_success "已创建 $USER_CONFIG"

        # 生成随机 secret_key
        local random_key
        random_key="$(openssl rand -hex 32 2>/dev/null || cat /dev/urandom | tr -dc 'a-zA-Z0-9' | head -c 64)"
        if [[ -n "$random_key" ]]; then
            # 替换 secret_key
            if [[ "$(uname -s)" == "Darwin" ]]; then
                sed -i '' "s/^secret_key =.*/secret_key = $random_key/" "$USER_CONFIG"
            else
                sed -i "s/^secret_key =.*/secret_key = $random_key/" "$USER_CONFIG"
            fi
            log_success "已自动生成 secret_key"
        else
            log_warn "无法自动生成 secret_key，请手动修改 $USER_CONFIG 中的 secret_key"
        fi

        # 默认使用 SQLite
        log_info "默认数据库类型: SQLite（数据库文件: data/${APP_NAME}.db）"
        log_info "如需使用 MySQL/PostgreSQL，请编辑 $USER_CONFIG 的 [database] 部分"

        echo ""
        print_separator
        log_warn ">>> 请检查 $USER_CONFIG 中的以下配置 <<<"
        echo ""
        echo "  1. [security] secret_key — 已自动生成，可自行修改"
        echo "  2. [server] http_port — 服务端口（默认 8080）"
        echo "  3. [server] domain   — 域名（默认 localhost）"
        echo "  4. [database]        — 数据库连接（默认 SQLite）"
        echo "  5. [mail]            — 邮件服务（如需找回密码）"
        echo ""
        print_separator
        echo ""

        # 给用户一个机会编辑配置
        log_warn "请确认配置无误后重新执行 ./deploy.sh 完成部署"
        return 2  # 特殊返回码：需要用户确认配置
    fi

    # 更新时：检查默认配置是否有新增项
    log_info "用户配置文件已存在，检查配置更新..."
    check_config_updates

    # 检查 secret_key 是否仍为默认值
    local secret_key
    secret_key="$(get_config_value "security" "secret_key")"
    if [[ -z "$secret_key" ]]; then
        log_error "secret_key 未设置！请在 $USER_CONFIG 的 [security] 部分设置 secret_key"
        return 1
    fi

    log_success "用户配置文件就绪: $USER_CONFIG"
    return 0
}

# 比较默认配置和用户配置，提示新增项
check_config_updates() {
    if [[ ! -f "$REPO_CONFIG" ]] || [[ ! -f "$USER_CONFIG" ]]; then
        return
    fi

    # 提取默认配置中的所有 section 和 key
    local repo_keys user_keys
    repo_keys="$(awk -F '=' '
        /^\[.*\]/ { sec = $0 }
        /^[a-z]/ && $1 !~ /^#/ { print sec "::" $1 }
    ' "$REPO_CONFIG" | sort)"

    user_keys="$(awk -F '=' '
        /^\[.*\]/ { sec = $0 }
        /^[a-z]/ && $1 !~ /^#/ { print sec "::" $1 }
    ' "$USER_CONFIG" | sort)"

    # 找出默认配置中有但用户配置中没有的项
    local new_items
    new_items="$(comm -23 <(echo "$repo_keys") <(echo "$user_keys") | head -20)"

    if [[ -n "$new_items" ]]; then
        log_warn "以下配置项在仓库默认配置中新增或变更，但你尚未同步到用户配置:"
        echo ""
        echo "$new_items" | while read -r item; do
            local sec_name="${item%%::*}"
            local key_name="${item##*::}"
            # 获取默认值
            local default_val
            default_val="$(awk -F '=' -v sec="$sec_name" -v k="$key_name" '
                $0 == sec { in_section=1; next }
                /^\[.*\]/ { in_section=0 }
                in_section && $1 ~ /^[[:space:]]*'"$key_name"'[[:space:]]*$/ {
                    val=$2; gsub(/^[[:space:]]+|[[:space:]]+$/, "", val); print val
                }
            ' "$REPO_CONFIG")"
            echo "  [$sec_name::$key_name] = $default_val"
        done
        echo ""
        log_info "请根据需要手动合并到 $USER_CONFIG"
    else
        log_info "用户配置与仓库默认配置一致"
    fi
}

# =============================================================================
# 构建
# =============================================================================

do_build_backend() {
    log_step "构建后端"

    # 清理旧缓存
    go clean -cache 2>/dev/null || true

    # 获取版本信息
    local version commit_hash
    version="$(grep '"version": ' package.json | awk -F ':' '{print $2}' | tr -d ' ",' || echo "unknown")"
    commit_hash="$(git rev-parse --short=7 HEAD 2>/dev/null || echo "unknown")"

    # 静态链接（Linux）
    local ldflags="-w -s"
    if [[ "$(uname -s)" == "Linux" ]]; then
        ldflags="$ldflags -linkmode external -extldflags '-static'"
    fi

    local build_args="-X main.Version=$version -X main.CommitHash=$commit_hash"

    log_info "版本: $version, Commit: $commit_hash"

    # 拉取依赖
    log_info "拉取 Go 依赖..."
    if ! go get . 2>&1; then
        log_error "go get 失败"
        return 1
    fi

    # 构建
    log_info "编译中..."
    if CGO_ENABLED=1 go build -a -v -trimpath \
        -ldflags "$ldflags $build_args" \
        -o "$NEW_BINARY" ezbookkeeping.go 2>&1; then
        chmod +x "$NEW_BINARY"
        log_success "后端构建完成 → $NEW_BINARY ($(du -h "$NEW_BINARY" | cut -f1))"
        return 0
    else
        log_error "后端构建失败"
        return 1
    fi
}

do_build_frontend() {
    log_step "构建前端"

    # 设置 Node 内存限制（低配服务器友好）
    export NODE_OPTIONS="--max-old-space-size=512"
    export NODE_ENV="production"

    # 安装依赖
    if [[ -d "node_modules" ]]; then
        log_info "更新依赖 (npm install)..."
    else
        log_info "安装依赖 (npm install)..."
    fi

    if ! npm install --prefer-offline --no-audit --no-fund 2>&1; then
        log_error "npm install 失败"
        return 1
    fi

    # 构建
    log_info "Vite 构建中（可能耗时数分钟，请耐心等待）..."
    if npm run build 2>&1; then
        log_success "前端构建完成 → $DIST_DIR/"
        return 0
    else
        log_error "前端构建失败"
        return 1
    fi
}

do_build() {
    local result=0

    if ! $SKIP_BACKEND; then
        if ! do_build_backend; then
            result=1
        fi
    else
        log_info "跳过后端构建 (--skip-backend)"
        if [[ ! -f "$BINARY_NAME" ]]; then
            log_error "--skip-backend 但 $BINARY_NAME 不存在，无法跳过首次后端构建"
            result=1
        fi
    fi

    if ! $SKIP_FRONTEND; then
        if ! do_build_frontend; then
            result=1
        fi
    else
        log_info "跳过前端构建 (--skip-frontend)"
        if [[ ! -d "$PUBLIC_DIR" ]]; then
            log_error "--skip-frontend 但 $PUBLIC_DIR/ 不存在，无法跳过首次前端构建"
            result=1
        fi
    fi

    return $result
}

# =============================================================================
# 服务管理
# =============================================================================

do_stop() {
    log_step "停止服务"

    # 优先使用 systemd
    if is_systemd_installed; then
        local status
        status="$(get_systemd_status)"
        if [[ "$status" == "active" ]]; then
            log_info "通过 systemd 停止服务..."
            if sudo systemctl stop "$SYSTEMD_SERVICE" 2>&1; then
                log_success "服务已停止 (systemd)"
                return 0
            else
                log_error "systemd 停止失败"
                return 1
            fi
        else
            log_info "服务未在运行 (systemd 状态: $status)"
            return 0
        fi
    fi

    # 回退：PID 文件 + 进程管理
    local pid
    pid="$(get_pid)"

    if [[ -z "$pid" ]]; then
        log_info "服务未在运行"
        rm -f "$PID_FILE"
        return 0
    fi

    log_info "向进程 $pid 发送 SIGTERM..."
    kill -TERM "$pid" 2>/dev/null || true

    # 等待优雅退出
    local waited=0
    while [[ "$waited" -lt "$GRACEFUL_TIMEOUT" ]]; do
        if ! kill -0 "$pid" 2>/dev/null; then
            log_success "服务已停止（PID: $pid, 耗时 ${waited}s）"
            rm -f "$PID_FILE"
            return 0
        fi
        sleep 1
        ((waited++))
    done

    # 超时，强制杀死
    log_warn "服务未在 ${GRACEFUL_TIMEOUT}s 内退出，发送 SIGKILL..."
    kill -KILL "$pid" 2>/dev/null || true
    sleep 1

    if kill -0 "$pid" 2>/dev/null; then
        log_error "无法停止进程 $pid"
        return 1
    fi

    log_success "服务已强制停止（PID: $pid）"
    rm -f "$PID_FILE"
    return 0
}

do_start() {
    log_step "启动服务"

    # 优先使用 systemd
    if is_systemd_installed; then
        local status
        status="$(get_systemd_status)"
        if [[ "$status" == "active" ]]; then
            log_warn "服务已在运行 (systemd)"
            log_info "如需重启请执行: ./deploy.sh restart"
            return 0
        fi
        log_info "通过 systemd 启动服务..."
        if sudo systemctl start "$SYSTEMD_SERVICE" 2>&1; then
            log_info "等待服务启动..."
            sleep 2
            if ! do_health_check; then
                log_error "服务启动后健康检查失败"
                return 1
            fi
            log_success "服务已启动 (systemd)"
            return 0
        else
            log_error "systemd 启动失败"
            return 1
        fi
    fi

    # 回退：nohup + PID 文件
    # 检查是否已在运行
    if is_running; then
        local pid
        pid="$(get_pid)"
        log_warn "服务已在运行（PID: $pid），跳过启动"
        log_info "如需重启请执行: ./deploy.sh restart"
        return 0
    fi

    # 确保必要的目录存在
    mkdir -p data log storage

    # 确定使用的二进制文件
    local binary="$BINARY_NAME"
    if [[ ! -f "$binary" ]]; then
        log_error "未找到 $BINARY_NAME，请先执行构建"
        return 1
    fi

    # 确定配置文件
    local config_file="$USER_CONFIG"
    if [[ ! -f "$config_file" ]]; then
        config_file="$REPO_CONFIG"
        log_warn "用户配置不存在，使用默认配置"
    fi

    local port
    port="$(get_server_port)"
    log_info "配置: --conf-path $config_file"
    log_info "端口: $port"

    # 启动服务
    nohup ./"$binary" server run --conf-path "$config_file" >> "log/${APP_NAME}.log" 2>&1 &
    local pid=$!
    echo "$pid" > "$PID_FILE"

    log_info "服务已启动（PID: $pid）"

    # 健康检查
    if ! do_health_check "$port"; then
        log_error "服务启动后健康检查失败"
        return 1
    fi

    log_success "服务启动成功 → http://127.0.0.1:${port}"
    return 0
}

do_restart() {
    print_separator
    log_info "开始重启 $(date '+%Y-%m-%d %H:%M:%S')"
    print_separator

    # 优先使用 systemd
    if is_systemd_installed; then
        # 替换文件
        if [[ -f "$NEW_BINARY" ]]; then
            if [[ -f "$BINARY_NAME" ]]; then
                cp "$BINARY_NAME" "$OLD_BINARY"
                log_info "旧二进制已备份 → $OLD_BINARY"
            fi
            mv "$NEW_BINARY" "$BINARY_NAME"
            log_success "新二进制已就位 → $BINARY_NAME"
        fi

        if [[ -d "$DIST_DIR" ]] && ! $SKIP_FRONTEND; then
            if [[ -d "$PUBLIC_DIR" ]]; then
                rm -rf "${PUBLIC_DIR}.old" 2>/dev/null
                mv "$PUBLIC_DIR" "${PUBLIC_DIR}.old"
                log_info "旧前端文件已备份 → ${PUBLIC_DIR}.old"
            fi
            mv "$DIST_DIR" "$PUBLIC_DIR"
            log_success "新前端文件已就位 → $PUBLIC_DIR/"
        fi

        log_info "通过 systemd 重启服务..."
        if sudo systemctl restart "$SYSTEMD_SERVICE" 2>&1; then
            sleep 2
            if do_health_check; then
                print_separator
                log_success "部署完成 ✓"
                print_separator
                return 0
            else
                log_error "服务重启后健康检查失败，执行回滚..."
                do_rollback
                return 1
            fi
        else
            log_error "systemd 重启失败"
            return 1
        fi
    fi

    # 回退：手动停止 + 替换 + 启动

    if ! do_stop; then
        log_error "停止旧服务失败，中止部署"
        return 1
    fi

    # 替换文件
    if [[ -f "$NEW_BINARY" ]]; then
        # 备份旧二进制
        if [[ -f "$BINARY_NAME" ]]; then
            cp "$BINARY_NAME" "$OLD_BINARY"
            log_info "旧二进制已备份 → $OLD_BINARY"
        fi
        mv "$NEW_BINARY" "$BINARY_NAME"
        log_success "新二进制已就位 → $BINARY_NAME"
    fi

    if [[ -d "$DIST_DIR" ]] && ! $SKIP_FRONTEND; then
        # 备份旧前端文件
        if [[ -d "$PUBLIC_DIR" ]]; then
            rm -rf "${PUBLIC_DIR}.old" 2>/dev/null
            mv "$PUBLIC_DIR" "${PUBLIC_DIR}.old"
            log_info "旧前端文件已备份 → ${PUBLIC_DIR}.old"
        fi
        mv "$DIST_DIR" "$PUBLIC_DIR"
        log_success "新前端文件已就位 → $PUBLIC_DIR/"
    fi

    if ! do_start; then
        log_error "新服务启动失败，执行回滚..."
        do_rollback
        return 1
    fi

    print_separator
    log_success "部署完成 ✓"
    print_separator
    return 0
}

# =============================================================================
# 健康检查
# =============================================================================

do_health_check() {
    local port="${1:-$(get_server_port)}"
    local elapsed=0

    log_info "健康检查: http://127.0.0.1:${port}/api/ (超时: ${HEALTH_CHECK_TIMEOUT}s)"

    while [[ "$elapsed" -lt "$HEALTH_CHECK_TIMEOUT" ]]; do
        local http_code
        http_code="$(curl -s -o /dev/null -w '%{http_code}' --connect-timeout 3 \
            "http://127.0.0.1:${port}/api/" 2>/dev/null || echo "000")"

        # 2xx/3xx/4xx 都表示服务已启动并响应（404 表示路由存在但 API 路径不对，也算存活）
        if [[ "$http_code" != "000" ]]; then
            log_success "服务响应正常（HTTP $http_code, 耗时 ${elapsed}s）"
            return 0
        fi

        sleep "$HEALTH_CHECK_INTERVAL"
        elapsed=$((elapsed + HEALTH_CHECK_INTERVAL))
        echo -n "."
    done

    echo ""
    log_error "服务在 ${HEALTH_CHECK_TIMEOUT}s 内未响应"
    return 1
}

# =============================================================================
# 回滚
# =============================================================================

do_rollback() {
    log_step "执行回滚"

    # 停止当前服务
    do_stop

    # 恢复旧二进制
    if [[ -f "$OLD_BINARY" ]]; then
        mv "$OLD_BINARY" "$BINARY_NAME"
        log_success "已恢复旧二进制"
    else
        log_warn "没有可回滚的旧二进制文件"
    fi

    # 恢复旧前端
    if [[ -d "${PUBLIC_DIR}.old" ]]; then
        rm -rf "$PUBLIC_DIR" 2>/dev/null
        mv "${PUBLIC_DIR}.old" "$PUBLIC_DIR"
        log_success "已恢复旧前端文件"
    fi

    # 清理未完成的构建产物
    rm -f "$NEW_BINARY"
    rm -rf "$DIST_DIR"

    # 尝试启动
    log_info "尝试启动回滚后的服务..."
    if do_start; then
        log_success "回滚成功，服务已恢复"
        return 0
    else
        log_error "回滚后服务未能启动，请手动排查"
        return 1
    fi
}

# =============================================================================
# 状态与日志查看
# =============================================================================

do_status() {
    # 优先使用 systemd
    if is_systemd_installed; then
        echo "管理方式: systemd"
        echo ""
        systemctl status "$SYSTEMD_SERVICE" 2>/dev/null || true
        return 0
    fi

    local pid
    pid="$(get_pid)"

    if [[ -z "$pid" ]]; then
        echo "状态: 未运行"
        rm -f "$PID_FILE"
        return 1
    fi

    echo "状态: 运行中"
    echo "PID:   $pid"

    # 进程信息
    if command -v ps &>/dev/null; then
        echo "进程:  $(ps -o pid,ppid,pcpu,rss,etime,command -p "$pid" 2>/dev/null | tail -1)"
    fi

    # 端口信息
    local port
    port="$(get_server_port)"
    echo "端口:  $port"

    # 磁盘使用
    if [[ -d "$PUBLIC_DIR" ]]; then
        echo "前端:  $(du -sh "$PUBLIC_DIR" 2>/dev/null | cut -f1)"
    fi
    if [[ -f "$BINARY_NAME" ]]; then
        echo "二进制: $(du -h "$BINARY_NAME" 2>/dev/null | cut -f1)"
    fi

    # 快速健康检查
    local http_code
    http_code="$(curl -s -o /dev/null -w '%{http_code}' --connect-timeout 3 \
        "http://127.0.0.1:${port}/api/" 2>/dev/null || echo "000")"
    if [[ "$http_code" != "000" ]]; then
        echo "健康:  ✓ (HTTP $http_code)"
    else
        echo "健康:  ✗ (无响应)"
    fi

    return 0
}

do_logs() {
    # 优先使用 systemd journal
    if is_systemd_installed; then
        echo "systemd 日志 ($SYSTEMD_SERVICE)，按 Ctrl+C 退出..."
        echo ""
        journalctl -u "$SYSTEMD_SERVICE" -f
        return 0
    fi

    local log_file="log/${APP_NAME}.log"

    if [[ ! -f "$log_file" ]]; then
        log_warn "日志文件 $log_file 不存在"
        log_info "也可以查看部署日志: tail -f $DEPLOY_LOG"
        return 1
    fi

    echo "实时日志 ($log_file)，按 Ctrl+C 退出..."
    echo ""
    tail -f "$log_file"
}

# =============================================================================
# systemd 服务安装
# =============================================================================

do_systemd_install() {
    log_step "安装 systemd 服务"

    if ! has_systemd; then
        log_error "当前系统不支持 systemd"
        return 1
    fi

    if is_systemd_installed; then
        log_info "systemd 服务文件已存在: $SYSTEMD_SERVICE_PATH"
        # 显示当前配置
        echo ""
        cat "$SYSTEMD_SERVICE_PATH"
        echo ""
        log_info "如需重新生成，请先删除: sudo rm $SYSTEMD_SERVICE_PATH"
        return 0
    fi

    # 检测运行用户
    local run_user="${SUDO_USER:-$USER}"
    if [[ "$run_user" == "root" ]]; then
        log_warn "以 root 运行，建议创建专用用户。将使用当前配置继续。"
    fi

    # 确定配置文件路径
    local config_path="${SCRIPT_DIR}/${USER_CONFIG}"
    if [[ ! -f "$config_path" ]]; then
        config_path="${SCRIPT_DIR}/${REPO_CONFIG}"
    fi

    log_info "运行用户: $run_user"
    log_info "工作目录: $SCRIPT_DIR"
    log_info "配置文件: $config_path"

    # 生成服务文件
    local tmp_service="/tmp/${SYSTEMD_SERVICE}"
    cat > "$tmp_service" << SERVICEOF
[Unit]
Description=ezbookkeeping - Enterprise Bookkeeping
Documentation=https://github.com/mayswind/ezbookkeeping
After=network.target

[Service]
Type=simple
User=$run_user
WorkingDirectory=$SCRIPT_DIR
ExecStart=$SCRIPT_DIR/$BINARY_NAME server run --conf-path $config_path
ExecStop=/bin/kill -SIGTERM \$MAINPID
Restart=always
RestartSec=5

# 安全加固
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=$SCRIPT_DIR/data $SCRIPT_DIR/log $SCRIPT_DIR/storage
ReadOnlyPaths=$SCRIPT_DIR/conf $SCRIPT_DIR/public $SCRIPT_DIR/templates

# 日志
StandardOutput=journal
StandardError=journal
SyslogIdentifier=$APP_NAME

[Install]
WantedBy=multi-user.target
SERVICEOF

    log_info "生成的服务文件内容:"
    echo ""
    echo "---"
    cat "$tmp_service"
    echo "---"
    echo ""

    # 安装需要 sudo
    log_info "需要 root 权限安装服务文件..."
    if ! sudo cp "$tmp_service" "$SYSTEMD_SERVICE_PATH"; then
        log_error "无法写入 $SYSTEMD_SERVICE_PATH"
        rm -f "$tmp_service"
        return 1
    fi
    rm -f "$tmp_service"

    if ! sudo systemctl daemon-reload; then
        log_error "systemctl daemon-reload 失败"
        return 1
    fi

    log_success "systemd 服务已安装: $SYSTEMD_SERVICE_PATH"

    echo ""
    print_separator
    log_info "后续操作:"
    echo ""
    echo "  # 启动服务"
    echo "  sudo systemctl start $SYSTEMD_SERVICE"
    echo ""
    echo "  # 开机自启"
    echo "  sudo systemctl enable $SYSTEMD_SERVICE"
    echo ""
    echo "  # 查看状态"
    echo "  systemctl status $SYSTEMD_SERVICE"
    echo ""
    echo "  # 查看日志"
    echo "  journalctl -u $SYSTEMD_SERVICE -f"
    echo ""
    echo "  # 之后每次部署会自动调用 systemctl restart"
    print_separator

    return 0
}

# =============================================================================
# Nginx 反向代理配置
# =============================================================================

do_nginx_setup() {
    log_step "配置 Nginx 反向代理"

    if ! command -v nginx &>/dev/null; then
        log_error "未找到 nginx，请先安装: sudo apt install nginx"
        return 1
    fi

    local nginx_dir
    nginx_dir="$(detect_nginx_config_dir)"
    if [[ "$nginx_dir" == "none" ]]; then
        log_error "未找到 Nginx 配置目录 (sites-available 或 conf.d)"
        return 1
    fi

    local config_file enable_cmd
    if [[ "$nginx_dir" == "sites" ]]; then
        config_file="$NGINX_SITES_AVAILABLE"
        enable_cmd="sudo ln -sf $NGINX_SITES_AVAILABLE $NGINX_SITES_ENABLED"
    else
        config_file="$NGINX_CONF_NGINX"
        enable_cmd=""
    fi

    local domain port
    port="$(get_server_port)"

    echo ""
    log_info "请输入以下信息（直接回车使用默认值）:"
    echo ""

    read -r -p "  域名 (如 ez.example.com): " domain
    if [[ -z "$domain" ]]; then
        log_error "域名不能为空"
        return 1
    fi

    local use_https
    read -r -p "  是否启用 HTTPS？[Y/n]: " use_https
    use_https="${use_https:-y}"
    if [[ "$use_https" =~ ^[Yy] ]]; then
        use_https="yes"
    else
        use_https="no"
    fi

    local cert_path="" key_path=""
    if [[ "$use_https" == "yes" ]]; then
        read -r -p "  SSL 证书路径 (fullchain.pem): " cert_path
        read -r -p "  SSL 私钥路径 (privkey.pem): " key_path
        if [[ -z "$cert_path" ]] || [[ -z "$key_path" ]]; then
            log_error "证书路径和私钥路径不能为空"
            return 1
        fi
    fi

    local tmp_conf="/tmp/nginx-${NGINX_CONF_NAME}.conf"
    local now_str
    now_str="$(date '+%Y-%m-%d %H:%M:%S')"

    # 生成配置头部
    cat > "$tmp_conf" << NGINXEOF
# ezbookkeeping Nginx 配置
# 由 deploy.sh 自动生成于 $now_str
# 域名: $domain

NGINXEOF

    # 通用反向代理 location 块
    local proxy_block
    proxy_block="    location / {
        proxy_pass http://127.0.0.1:$port;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;

        proxy_connect_timeout 60s;
        proxy_send_timeout    60s;
        proxy_read_timeout    60s;

        client_max_body_size 20m;
    }"

    if [[ "$use_https" == "yes" ]]; then
        cat >> "$tmp_conf" << NGINXEOF
# HTTP → HTTPS 重定向
server {
    listen 80;
    server_name $domain;

    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    location / {
        return 301 https://\$host\$request_uri;
    }
}

server {
    listen 443 ssl http2;
    server_name $domain;

    ssl_certificate     $cert_path;
    ssl_certificate_key $key_path;
    ssl_protocols       TLSv1.2 TLSv1.3;
    ssl_ciphers         HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    ssl_session_cache   shared:SSL:10m;
    ssl_session_timeout 10m;

    add_header Strict-Transport-Security "max-age=31536000" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;

$proxy_block
}
NGINXEOF
    else
        cat >> "$tmp_conf" << NGINXEOF
server {
    listen 80;
    server_name $domain;

$proxy_block
}
NGINXEOF
    fi

    log_info "生成的 Nginx 配置:"
    echo ""
    echo "---"
    cat "$tmp_conf"
    echo "---"
    echo ""

    read -r -p "  是否安装此配置？[Y/n]: " confirm
    confirm="${confirm:-y}"
    if [[ ! "$confirm" =~ ^[Yy] ]]; then
        log_info "已取消"
        rm -f "$tmp_conf"
        return 0
    fi

    log_info "需要 root 权限安装 Nginx 配置..."
    if ! sudo cp "$tmp_conf" "$config_file"; then
        log_error "无法写入 $config_file"
        rm -f "$tmp_conf"
        return 1
    fi
    rm -f "$tmp_conf"

    if [[ -n "$enable_cmd" ]]; then
        eval "$enable_cmd" || log_warn "创建软链接失败"
    fi

    if ! sudo nginx -t 2>&1; then
        log_error "Nginx 配置测试失败，请检查"
        return 1
    fi

    if ! sudo nginx -s reload 2>&1; then
        log_error "Nginx 重载失败"
        return 1
    fi

    log_success "Nginx 配置已安装并重载"
    echo ""
    if [[ "$use_https" == "yes" ]]; then
        log_info "访问地址: https://$domain"
        echo ""
        log_info "=== 免费 Let's Encrypt 证书（如尚未配置）==="
        echo ""
        echo "  sudo apt install certbot python3-certbot-nginx"
        echo "  sudo certbot --nginx -d $domain"
        echo ""
        echo "  certbot 会自动修改 Nginx 配置并设置自动续期"
    else
        log_info "访问地址: http://$domain"
    fi

    return 0
}

do_deploy() {
    # 轮换部署日志
    rotate_deploy_log

    echo ""
    print_separator
    log_info "ezbookkeeping 部署脚本"
    log_info "时间: $(date '+%Y-%m-%d %H:%M:%S')"
    log_info "目录: $SCRIPT_DIR"
    print_separator
    echo ""

    # 1. 环境检查
    do_check || return 1

    # 2. 代码更新
    do_pull || return 1

    # 3. 配置初始化
    local config_result
    do_config_init
    config_result=$?
    if [[ "$config_result" -eq 2 ]]; then
        # 首次部署，需要用户确认配置
        log_info "配置初始化完成，请检查配置后再次执行部署"
        return 0
    elif [[ "$config_result" -ne 0 ]]; then
        return 1
    fi

    # 4. 构建
    do_build || return 1

    # 5. 重启服务
    do_restart || return 1

    # 6. 部署后建议
    echo ""
    local suggestions=()
    if has_systemd && ! is_systemd_installed; then
        suggestions+=("systemd")
    fi
    if command -v nginx &>/dev/null && [[ "$(detect_nginx_config_dir)" != "none" ]]; then
        suggestions+=("nginx")
    fi

    if [[ ${#suggestions[@]} -gt 0 ]]; then
        print_separator
        log_info "建议执行的后续操作:"
        echo ""
        for s in "${suggestions[@]}"; do
            case "$s" in
                systemd)
                    echo "  ./deploy.sh systemd              # 安装 systemd 服务（崩溃自动重启、开机自启）"
                    ;;
                nginx)
                    echo "  ./deploy.sh nginx                # 配置 Nginx 反向代理（HTTPS、静态文件加速）"
                    ;;
            esac
        done
        print_separator
    fi

    return 0
}

# =============================================================================
# 帮助
# =============================================================================

show_help() {
    cat << EOF
ezbookkeeping 部署脚本

用法:
  ./deploy.sh               完整部署（环境检查 → 拉代码 → 构建 → 重启）
  ./deploy.sh start         启动服务
  ./deploy.sh stop          停止服务
  ./deploy.sh restart       重启服务
  ./deploy.sh status        查看服务状态
  ./deploy.sh logs          查看服务实时日志
  ./deploy.sh build         仅构建（不重启）
  ./deploy.sh check         仅环境检查
  ./deploy.sh rollback      回滚到上一个版本
  ./deploy.sh systemd       安装 systemd 服务（崩溃自动重启、开机自启）
  ./deploy.sh nginx         配置 Nginx 反向代理（HTTPS、静态文件加速）

选项:
  --skip-frontend           跳过前端构建
  --skip-backend            跳过后端构建
  -h, --help                显示此帮助

示例:
  ./deploy.sh                          # 完整部署
  ./deploy.sh --skip-frontend          # 只更新后端
  ./deploy.sh restart                  # 仅重启（不拉代码、不构建）
  ./deploy.sh status                   # 查看运行状态
  ./deploy.sh systemd                  # 生成 systemd 服务文件

首次部署（推荐流程）:
  1. 上传代码到服务器
  2. ./deploy.sh check                 # 检查环境
  3. ./deploy.sh                       # 首次部署（自动创建配置）
  4. 检查 conf/ezbookkeeping.user.ini，确认无误
  5. ./deploy.sh                       # 再次执行完成部署
  6. ./deploy.sh systemd               # 安装 systemd 服务
     sudo systemctl enable --now ezbookkeeping
  7. ./deploy.sh nginx                 # (可选) 配置 Nginx 反向代理

日常更新:
  ./deploy.sh              # git pull → 构建 → systemctl restart

配置文件:
  仓库模板:    conf/ezbookkeeping.ini        (git 跟踪，随版本更新)
  用户配置:    conf/ezbookkeeping.user.ini   (首次自动创建，永不覆盖)
EOF
}

# =============================================================================
# 入口
# =============================================================================

# 解析参数
COMMAND=""
while [[ $# -gt 0 ]]; do
    case "$1" in
        start|stop|restart|status|logs|build|check|rollback|systemd|nginx)
            COMMAND="$1"
            ;;
        --skip-frontend)
            SKIP_FRONTEND=true
            ;;
        --skip-backend)
            SKIP_BACKEND=true
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            log_error "未知参数: $1"
            echo ""
            show_help
            exit 2
            ;;
    esac
    shift
done

# 确保在项目根目录
if [[ ! -f "ezbookkeeping.go" ]]; then
    log_error "请在项目根目录下执行此脚本（ezbookkeeping.go 所在目录）"
    exit 1
fi

# 执行命令
case "$COMMAND" in
    start)
        do_start
        ;;
    stop)
        do_stop
        ;;
    restart)
        do_restart
        ;;
    status)
        do_status
        ;;
    logs)
        do_logs
        ;;
    build)
        do_check && do_build
        ;;
    check)
        do_check
        ;;
    rollback)
        do_rollback
        ;;
    systemd)
        do_systemd_install
        ;;
    nginx)
        do_nginx_setup
        ;;
    *)
        # 默认：完整部署
        do_deploy
        ;;
esac

exit $?
