#!/bin/bash
# ============================================================
# 生猪买卖测试数据种子脚本
# 用法: bash scripts/seed_pig_trading.sh [后端地址]
# 默认后端地址: http://localhost:8080
#
# 前置条件: 后端服务已启动
# ============================================================

set -euo pipefail

BASE_URL="${1:-http://localhost:8081}"
API="$BASE_URL/api"
API_V1="$API/v1"

# 测试用户凭据
USERNAME="pig_trader"
EMAIL="pig_trader@test.com"
PASSWORD="password123"
NICKNAME="猪老板"

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

info()  { printf "${BLUE}[INFO]${NC}  %s\n" "$*"; }
ok()    { printf "${GREEN}[OK]${NC}    %s\n" "$*"; }
err()   { printf "${RED}[ERR]${NC}   %s\n" "$*"; exit 1; }

# --- 检查依赖 ---
check_deps() {
    command -v curl >/dev/null 2>&1 || err "需要 curl，请先安装"
    command -v jq   >/dev/null 2>&1 || err "需要 jq，请先安装 (brew install jq)"
}

# --- 检查后端可用 ---
wait_backend() {
    info "等待后端服务就绪: $BASE_URL"
    local max=30
    for ((i=1; i<=max; i++)); do
        if curl -s --connect-timeout 2 "$BASE_URL/api/authorize.json" >/dev/null 2>&1; then
            ok "后端服务已就绪"
            return 0
        fi
        sleep 1
    done
    err "后端服务未启动 ($BASE_URL)，请先执行 'go run . server run --conf-path conf/ezbookkeeping.dev.ini'"
}

# --- 注册用户 ---
register_user() {
    info "注册测试用户: $USERNAME"
    local resp
    resp=$(curl -s -X POST "$API/register.json" \
        -H "Content-Type: application/json" \
        -d '{
            "username":"'"$USERNAME"'",
            "email":"'"$EMAIL"'",
            "nickname":"'"$NICKNAME"'",
            "password":"'"$PASSWORD"'",
            "language":"zh_CN",
            "defaultCurrency":"CNY",
            "firstDayOfWeek":0,
            "categories":[]
        }' || true)

    TOKEN=$(echo "$resp" | jq -r '.result.token // empty')
    if [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
        ok "注册成功，已获取 token"
    else
        # 可能用户已存在，尝试登录
        err_msg=$(echo "$resp" | jq -r '.errorMessage // empty' 2>/dev/null || true)
        if echo "$err_msg" | grep -qi "exist\|exists\|already\|already"; then
            info "用户已存在，跳过注册"
        else
            info "注册返回: $resp"
            info "尝试直接登录..."
        fi
    fi
}

# --- 登录获取 token ---
login_user() {
    if [ -n "${TOKEN:-}" ] && [ "$TOKEN" != "null" ]; then
        return 0
    fi
    info "登录获取 token..."
    local resp
    resp=$(curl -s -X POST "$API/authorize.json" \
        -H "Content-Type: application/json" \
        -d '{
            "loginName":"'"$USERNAME"'",
            "password":"'"$PASSWORD"'"
        }' || true)

    # 检查是否需要 2FA
    local need_2fa
    need_2fa=$(echo "$resp" | jq -r '.result.need2FA // false')
    if [ "$need_2fa" = "true" ]; then
        err "需要 2FA 认证，请确认配置中 enable_two_factor = false"
    fi

    TOKEN=$(echo "$resp" | jq -r '.result.token // empty')
    if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
        err "登录失败: $resp"
    fi
    ok "登录成功，已获取 token"
}

H_AUTH=""
H_TZ="X-Timezone-Offset: 480"

auth_headers() {
    H_AUTH="Authorization: Bearer $TOKEN"
}

# --- 创建交易分类 ---
create_categories() {
    info "创建生猪买卖相关分类..."

    local categories_json
    categories_json=$(cat <<'JSON'
{"categories":[
    {
        "name":"生猪买卖",
        "type":1,
        "icon":"2669293",
        "color":"4CAF50",
        "subCategories":[
            {"name":"生猪销售","type":1,"icon":"1518688","color":"66BB6A","parentId":"0"},
            {"name":"仔猪销售","type":1,"icon":"1518688","color":"81C784","parentId":"0"}
        ]
    },
    {
        "name":"饲料采购",
        "type":2,
        "icon":"2994559",
        "color":"FF9800",
        "subCategories":[
            {"name":"玉米","type":2,"icon":"2994559","color":"FFB74D","parentId":"0"},
            {"name":"豆粕","type":2,"icon":"2994559","color":"FFCC80","parentId":"0"},
            {"name":"预混料","type":2,"icon":"2994559","color":"FFE0B2","parentId":"0"}
        ]
    },
    {
        "name":"仔猪采购",
        "type":2,
        "icon":"2994559",
        "color":"F44336",
        "subCategories":[
            {"name":"仔猪购买","type":2,"icon":"2994559","color":"EF5350","parentId":"0"}
        ]
    },
    {
        "name":"疫苗药品",
        "type":2,
        "icon":"2656383",
        "color":"9C27B0",
        "subCategories":[
            {"name":"疫苗","type":2,"icon":"2656383","color":"BA68C8","parentId":"0"},
            {"name":"兽药","type":2,"icon":"2656383","color":"CE93D8","parentId":"0"}
        ]
    },
    {
        "name":"猪舍维护",
        "type":2,
        "icon":"3037320",
        "color":"795548",
        "subCategories":[
            {"name":"猪舍维修","type":2,"icon":"3037320","color":"8D6E63","parentId":"0"}
        ]
    },
    {
        "name":"运输费用",
        "type":2,
        "icon":"1575070",
        "color":"607D8B",
        "subCategories":[
            {"name":"货物运输","type":2,"icon":"1575070","color":"78909C","parentId":"0"}
        ]
    },
    {
        "name":"屠宰加工",
        "type":2,
        "icon":"2983881",
        "color":"455A64",
        "subCategories":[
            {"name":"屠宰加工费","type":2,"icon":"2983881","color":"546E7A","parentId":"0"}
        ]
    },
    {
        "name":"水电杂费",
        "type":2,
        "icon":"2938705",
        "color":"00BCD4",
        "subCategories":[
            {"name":"水费","type":2,"icon":"2938705","color":"80DEEA","parentId":"0"},
            {"name":"电费","type":2,"icon":"2938705","color":"B2EBF2","parentId":"0"}
        ]
    },
    {
        "name":"人工工资",
        "type":2,
        "icon":"2635049",
        "color":"E91E63",
        "subCategories":[
            {"name":"工资发放","type":2,"icon":"2635049","color":"F06292","parentId":"0"}
        ]
    },
    {
        "name":"设备采购",
        "type":2,
        "icon":"2710592",
        "color":"3F51B5",
        "subCategories":[
            {"name":"设备购置","type":2,"icon":"2710592","color":"5C6BC0","parentId":"0"}
        ]
    }
]}
JSON
)

    local resp
    resp=$(curl -s -X POST "$API_V1/transaction/categories/add_batch.json" \
        -H "Content-Type: application/json" \
        -H "$H_AUTH" -H "$H_TZ" \
        -d "$categories_json")

    local saved
    saved=$(echo "$resp" | jq -r '.success // false')
    if [ "$saved" = "true" ]; then
        ok "分类创建成功"
    else
        err_msg=$(echo "$resp" | jq -r '.errorMessage // empty' 2>/dev/null || true)
        if echo "$err_msg" | grep -qi "already\|exist\|duplicate"; then
            ok "分类已存在，跳过"
        else
            echo "$resp" | jq . 2>/dev/null || echo "$resp"
            ok "分类创建完成（部分可能已存在）"
        fi
    fi
}

# --- 创建账户 ---
create_accounts() {
    info "创建账户..."

    local accounts=(
        '{"name":"现金钱包","category":1,"type":1,"icon":"2925141","color":"4CAF50","currency":"CNY","balance":1000000,"balanceTime":1772294400,"comment":"日常现金"}'
        '{"name":"农业银行卡","category":2,"type":1,"icon":"1587533","color":"2196F3","currency":"CNY","balance":50000000,"balanceTime":1772294400,"comment":"主要账户"}'
        '{"name":"应收账款-猪肉批发商","category":6,"type":1,"icon":"2691791","color":"FF5722","currency":"CNY","balance":2000000,"balanceTime":1772294400,"comment":"赊账销售款项"}'
        '{"name":"应付账款-饲料供应商","category":5,"type":1,"icon":"2691791","color":"9C27B0","currency":"CNY","balance":-800000,"balanceTime":1772294400,"comment":"欠饲料款"}'
    )

    for acc_json in "${accounts[@]}"; do
        local name
        name=$(echo "$acc_json" | jq -r '.name')
        info "  创建账户: $name"
        local resp
        resp=$(curl -s -X POST "$API_V1/accounts/add.json" \
            -H "Content-Type: application/json" \
            -H "$H_AUTH" -H "$H_TZ" \
            -d "$acc_json") || true

        local uid
        uid=$(echo "$resp" | jq -r '.result.id // empty' 2>/dev/null || true)
        if [ -n "$uid" ]; then
            ok "    已创建 (id=$uid)"
        else
            info "    可能已存在，跳过"
        fi
    done
}

# --- 获取已有账户 ID ---
get_account_ids() {
    info "获取账户列表..."
    local resp
    resp=$(curl -s "$API_V1/accounts/list.json" \
        -H "$H_AUTH" -H "$H_TZ" | jq '.')

    CASH_ID=$(echo "$resp" | jq -r '.result[] | select(.name=="现金钱包") | .id' 2>/dev/null | head -1 || echo "")
    BANK_ID=$(echo "$resp" | jq -r '.result[] | select(.name=="农业银行卡") | .id' 2>/dev/null | head -1 || echo "")
    RECV_ID=$(echo "$resp" | jq -r '.result[] | select(.name=="应收账款-猪肉批发商") | .id' 2>/dev/null | head -1 || echo "")
    PAY_ID=$(echo "$resp" | jq -r '.result[] | select(.name=="应付账款-饲料供应商") | .id' 2>/dev/null | head -1 || echo "")

    if [ -z "$BANK_ID" ]; then
        err "未找到农业银行卡账户"
    fi
    ok "已获取账户 ID: 现金=$CASH_ID, 银行卡=$BANK_ID, 应收=$RECV_ID, 应付=$PAY_ID"
}

# --- 获取分类 ID ---
# 在分类列表中按名称查找（支持父分类和子分类）
find_cat_id() {
    local name=$1 resp=$2
    echo "$resp" | jq -r --arg n "$name" '
        [.result["1"][], .result["2"][]]
        | .[] | . as $p | ($p, ($p.subCategories[]?))
        | select(.name==$n) | .id
    ' 2>/dev/null | head -1
}

get_category_ids() {
    info "获取分类列表..."
    local resp
    resp=$(curl -s "$API_V1/transaction/categories/list.json" \
        -H "$H_AUTH" -H "$H_TZ")

    CAT_PIG_SALE=$(find_cat_id "生猪销售" "$resp")
    CAT_PIGLET_SALE=$(find_cat_id "仔猪销售" "$resp")

    CAT_PIGLET_BUY=$(find_cat_id "仔猪购买" "$resp")
    CAT_FEED_CORN=$(find_cat_id "玉米" "$resp")
    CAT_FEED_SOYMEAL=$(find_cat_id "豆粕" "$resp")
    CAT_FEED_PREMIX=$(find_cat_id "预混料" "$resp")
    CAT_VACCINE=$(find_cat_id "疫苗" "$resp")
    CAT_MEDICINE=$(find_cat_id "兽药" "$resp")
    CAT_MAINTENANCE=$(find_cat_id "猪舍维修" "$resp")
    CAT_TRANSPORT=$(find_cat_id "货物运输" "$resp")
    CAT_SLAUGHTER=$(find_cat_id "屠宰加工费" "$resp")
    CAT_WATER=$(find_cat_id "水费" "$resp")
    CAT_ELECTRIC=$(find_cat_id "电费" "$resp")
    CAT_WAGE=$(find_cat_id "工资发放" "$resp")
    CAT_EQUIPMENT=$(find_cat_id "设备购置" "$resp")

    ok "已获取分类 ID"
}

# --- 创建交易 ---
create_transactions() {
    info "创建测试交易记录..."

    set +e

    local now
    now=$(date +%s)

    # UTC+8 offset
    local tz=480

    # 辅助函数: 将 YYYY-MM-DD 转为 Unix timestamp
    date_to_ts() {
        # macOS date
        date -j -f "%Y-%m-%dT%H:%M:%S" "$1T12:00:00" +%s 2>/dev/null || \
        date -d "$1 12:00:00" +%s 2>/dev/null
    }

    # 使用固定时间戳（北京时间 2026年3月-5月）
    local t1=$(date_to_ts "2026-03-15")
    local t2=$(date_to_ts "2026-03-20")
    local t3=$(date_to_ts "2026-03-25")
    local t4=$(date_to_ts "2026-04-01")
    local t5=$(date_to_ts "2026-04-08")
    local t6=$(date_to_ts "2026-04-12")
    local t7=$(date_to_ts "2026-04-18")
    local t8=$(date_to_ts "2026-04-25")
    local t9=$(date_to_ts "2026-05-01")
    local t10=$(date_to_ts "2026-05-05")
    local t11=$(date_to_ts "2026-05-10")
    local t12=$(date_to_ts "2026-05-15")
    local t13=$(date_to_ts "2026-05-17")
    local t14=$(date_to_ts "2026-05-17")

    local transactions=()
    local count=0

    add_tx() {
        transactions+=("$1")
        ((count++))
    }

    # --- 支出类交易（从银行卡出） ---

    # 1. 购买仔猪 50头 × 500元 = 25,000元
    add_tx '{"type":3,"categoryId":"'"$CAT_PIGLET_BUY"'","time":'"$t1"',"utcOffset":'"$tz"',"sourceAccountId":"'"$BANK_ID"'","sourceAmount":2500000,"hideAmount":false,"tagIds":[],"comment":"购买仔猪 50头 × 500元/头"}'

    # 2. 采购玉米 2吨 × 2800元 = 5,600元
    add_tx '{"type":3,"categoryId":"'"$CAT_FEED_CORN"'","time":'"$t2"',"utcOffset":'"$tz"',"sourceAccountId":"'"$BANK_ID"'","sourceAmount":560000,"hideAmount":false,"tagIds":[],"comment":"采购玉米 2吨"}'

    # 3. 采购豆粕 1.5吨 × 4200元 = 6,300元
    add_tx '{"type":3,"categoryId":"'"$CAT_FEED_SOYMEAL"'","time":'"$t2"',"utcOffset":'"$tz"',"sourceAccountId":"'"$BANK_ID"'","sourceAmount":630000,"hideAmount":false,"tagIds":[],"comment":"采购豆粕 1.5吨"}'

    # 4. 猪瘟疫苗 一批 = 3,500元
    add_tx '{"type":3,"categoryId":"'"$CAT_VACCINE"'","time":'"$t3"',"utcOffset":'"$tz"',"sourceAccountId":"'"$BANK_ID"'","sourceAmount":350000,"hideAmount":false,"tagIds":[],"comment":"猪瘟疫苗 500头份"}'

    # 5. 兽药（抗生素）一批 = 1,800元
    add_tx '{"type":3,"categoryId":"'"$CAT_MEDICINE"'","time":'"$t3"',"utcOffset":'"$tz"',"sourceAccountId":"'"$BANK_ID"'","sourceAmount":180000,"hideAmount":false,"tagIds":[],"comment":"抗生素及驱虫药"}'

    # 6. 预混料 0.5吨 × 6000元 = 3,000元
    add_tx '{"type":3,"categoryId":"'"$CAT_FEED_PREMIX"'","time":'"$t4"',"utcOffset":'"$tz"',"sourceAccountId":"'"$BANK_ID"'","sourceAmount":300000,"hideAmount":false,"tagIds":[],"comment":"预混料 0.5吨"}'

    # 7. 运输费 运猪车一趟 = 1,200元
    add_tx '{"type":3,"categoryId":"'"$CAT_TRANSPORT"'","time":'"$t5"',"utcOffset":'"$tz"',"sourceAccountId":"'"$CASH_ID"'","sourceAmount":120000,"hideAmount":false,"tagIds":[],"comment":"生猪运输 至批发市场"}'

    # 8. 猪舍围栏维修 = 2,000元
    add_tx '{"type":3,"categoryId":"'"$CAT_MAINTENANCE"'","time":'"$t6"',"utcOffset":'"$tz"',"sourceAccountId":"'"$BANK_ID"'","sourceAmount":200000,"hideAmount":false,"tagIds":[],"comment":"更换猪舍围栏"}'

    # 9. 水电费 5月 = 3,200元
    add_tx '{"type":3,"categoryId":"'"$CAT_ELECTRIC"'","time":'"$t7"',"utcOffset":'"$tz"',"sourceAccountId":"'"$BANK_ID"'","sourceAmount":320000,"hideAmount":false,"tagIds":[],"comment":"5月份电费"}'

    # --- 收入类交易 ---

    # 10. 出售生猪 20头 × 2500元 = 50,000元
    add_tx '{"type":2,"categoryId":"'"$CAT_PIG_SALE"'","time":'"$t4"',"utcOffset":'"$tz"',"sourceAccountId":"'"$BANK_ID"'","sourceAmount":5000000,"hideAmount":false,"tagIds":[],"comment":"出售生猪 20头 × 2500元/头"}'

    # 11. 出售仔猪 10头 × 800元 = 8,000元
    add_tx '{"type":2,"categoryId":"'"$CAT_PIGLET_SALE"'","time":'"$t8"',"utcOffset":'"$tz"',"sourceAccountId":"'"$CASH_ID"'","sourceAmount":800000,"hideAmount":false,"tagIds":[],"comment":"出售仔猪 10头"}'

    # 12. 出售生猪 15头 × 2800元 = 42,000元
    add_tx '{"type":2,"categoryId":"'"$CAT_PIG_SALE"'","time":'"$t9"',"utcOffset":'"$tz"',"sourceAccountId":"'"$BANK_ID"'","sourceAmount":4200000,"hideAmount":false,"tagIds":[],"comment":"出售生猪 15头 × 2800元/头"}'

    # --- 更多支出 ---

    # 13. 屠宰加工费 20头 × 150元 = 3,000元
    add_tx '{"type":3,"categoryId":"'"$CAT_SLAUGHTER"'","time":'"$t10"',"utcOffset":'"$tz"',"sourceAccountId":"'"$BANK_ID"'","sourceAmount":300000,"hideAmount":false,"tagIds":[],"comment":"生猪屠宰加工 20头"}'

    # 14. 饲料采购（玉米）3吨 × 2800元 = 8,400元
    add_tx '{"type":3,"categoryId":"'"$CAT_FEED_CORN"'","time":'"$t11"',"utcOffset":'"$tz"',"sourceAccountId":"'"$BANK_ID"'","sourceAmount":840000,"hideAmount":false,"tagIds":[],"comment":"采购玉米 3吨（补库存）"}'

    # 15. 人工工资 6月 = 18,000元
    add_tx '{"type":3,"categoryId":"'"$CAT_WAGE"'","time":'"$t12"',"utcOffset":'"$tz"',"sourceAccountId":"'"$BANK_ID"'","sourceAmount":1800000,"hideAmount":false,"tagIds":[],"comment":"6月份工人工资（3人）"}'

    # 16. 出售生猪 25头 × 2600元 = 65,000元
    add_tx '{"type":2,"categoryId":"'"$CAT_PIG_SALE"'","time":'"$t13"',"utcOffset":'"$tz"',"sourceAccountId":"'"$BANK_ID"'","sourceAmount":6500000,"hideAmount":false,"tagIds":[],"comment":"出售生猪 25头 × 2600元/头"}'

    # 17. 水费 = 800元
    add_tx '{"type":3,"categoryId":"'"$CAT_WATER"'","time":'"$t14"',"utcOffset":'"$tz"',"sourceAccountId":"'"$CASH_ID"'","sourceAmount":80000,"hideAmount":false,"tagIds":[],"comment":"6月份水费"}'

    # 18. 采购自动喂料机 = 35,000元
    add_tx '{"type":3,"categoryId":"'"$CAT_EQUIPMENT"'","time":'"$t14"',"utcOffset":'"$tz"',"sourceAccountId":"'"$BANK_ID"'","sourceAmount":3500000,"hideAmount":false,"tagIds":[],"comment":"采购自动喂料机 2台"}'

    # --- 逐个发送 ---
    local idx=1 success=0 fail=0
    for tx_json in "${transactions[@]}"; do
        local comment
        comment=$(echo "$tx_json" | jq -r '.comment')
        info "  创建交易 [$idx/$count]: $comment"

        local resp
        resp=$(curl -s -X POST "$API_V1/transactions/add.json" \
            -H "Content-Type: application/json" \
            -H "$H_AUTH" -H "$H_TZ" \
            -d "$tx_json") || true

        local txid
        txid=$(echo "$resp" | jq -r '.result.id // empty' 2>/dev/null || true)
        if [ -n "$txid" ] && [ "$txid" != "null" ]; then
            ok "    已创建 (id=$txid)"
            ((success++))
        else
            local err_msg
            err_msg=$(echo "$resp" | jq -r '.errorMessage // empty' 2>/dev/null || true)
            info "    失败: ${err_msg:-$resp}"
            ((fail++))
        fi
        ((idx++))
    done

    TX_SUCCESS=$success
    TX_FAIL=$fail
    set -e
}

# --- 主流程 ---
main() {
    TX_SUCCESS=0
    TX_FAIL=0

    echo ""
    echo "============================================================"
    echo "  生猪买卖测试数据种子脚本"
    echo "============================================================"
    echo ""

    check_deps
    wait_backend
    register_user
    login_user
    auth_headers
    create_categories
    create_accounts
    get_account_ids
    get_category_ids
    create_transactions

    echo ""
    echo "============================================================"
    if [ "$TX_FAIL" -eq 0 ]; then
        ok "全部测试数据插入完成！（成功: $TX_SUCCESS 条）"
    else
        info "测试数据插入完成（成功: $TX_SUCCESS 条, 失败: $TX_FAIL 条）"
    fi
    echo ""
    echo "  登录信息:"
    echo "    用户名: $USERNAME"
    echo "    密码:   $PASSWORD"
    echo ""
    echo "  账户:"
    echo "    现金钱包 (Cash)"
    echo "    农业银行卡 (Checking)"
    echo "    应收账款-猪肉批发商 (Receivables)"
    echo "    应付账款-饲料供应商 (Debt)"
    echo ""
    echo "  共创建 18 条交易记录，涵盖："
    echo "    - 仔猪采购 / 生猪销售 / 仔猪销售"
    echo "    - 饲料采购（玉米、豆粕、预混料）"
    echo "    - 疫苗药品 / 猪舍维护"
    echo "    - 运输费用 / 屠宰加工"
    echo "    - 水电杂费 / 人工工资"
    echo "    - 设备采购"
    echo "============================================================"
    echo ""
}

main
