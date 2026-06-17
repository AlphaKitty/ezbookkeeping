/**
 * A field that may have a computed expression.
 */
export interface FieldExpr {
    key: string;
    expr: string;
}

/**
 * Evaluate all computed fields in dependency order.
 *
 * Fields whose dependencies are unsatisfied are silently omitted from the result
 * (empty-on-incomplete). Returns a map from field key to computed value.
 */
export function evaluateFieldExpressions(
    fields: FieldExpr[],
    values: Record<string, number>,
): Record<string, number> {
    if (fields.length === 0) {
        return {};
    }

    // Copy values so we don't mutate the input
    const allValues: Record<string, number> = { ...values };
    const result: Record<string, number> = {};

    // Build dependency graph
    const exprMap: Record<string, string> = {};
    const indegree: Record<string, number> = {};
    const dependents: Record<string, string[]> = {};

    for (const f of fields) {
        exprMap[f.key] = f.expr;
    }

    for (const f of fields) {
        const deps = extractDeps(f.expr, exprMap);
        indegree[f.key] = deps.length;
        for (const dep of deps) {
            if (!dependents[dep]) {
                dependents[dep] = [];
            }
            dependents[dep].push(f.key);
        }
    }

    // Kahn's algorithm
    const queue: string[] = [];
    for (const f of fields) {
        if (indegree[f.key] === 0) {
            queue.push(f.key);
        }
    }

    while (queue.length > 0) {
        const key = queue.shift()!;
        const expr = exprMap[key]!;

        try {
            const val = evaluateExpression(expr, allValues);
            allValues[key] = val;
            result[key] = val;
        } catch {
            // Dependency not satisfied — skip (empty-on-incomplete)
            continue;
        }

        for (const dep of dependents[key] || []) {
            indegree[dep]!--;
            if (indegree[dep] === 0) {
                queue.push(dep);
            }
        }
    }

    return result;
}

/**
 * Extract variable names from an expression that reference other computed fields.
 */
function extractDeps(expr: string, exprMap: Record<string, string>): string[] {
    const deps: string[] = [];
    const seen = new Set<string>();

    const re = /[a-zA-Z_][a-zA-Z0-9_]*/g;
    let match: RegExpExecArray | null;
    while ((match = re.exec(expr)) !== null) {
        const name = match[0];
        if (name in exprMap && !seen.has(name)) {
            deps.push(name);
            seen.add(name);
        }
    }

    return deps;
}

/**
 * Evaluate a simple arithmetic expression with variables.
 *
 * Supports: +, -, *, /, parentheses, variable names (letters/underscore identifiers),
 * unary minus, and decimal numbers.
 *
 * Grammar matches the Go `EvaluateExpression` in `pkg/utils/expression.go`.
 */
export function evaluateExpression(expr: string, vars: Record<string, number>): number {
    expr = expr.trim();
    if (expr === '') {
        return 0;
    }

    const tokens = tokenize(expr, vars);
    const parser = new Parser(tokens);
    const result = parser.parseExpression();

    if (parser.pos < parser.tokens.length) {
        throw new Error('unexpected token after expression');
    }

    return result;
}

enum TokenType {
    Number = 0,
    Plus,
    Minus,
    Star,
    Slash,
    LParen,
    RParen,
}

interface Token {
    type: TokenType;
    value: number;
}

function tokenize(expr: string, vars: Record<string, number>): Token[] {
    const tokens: Token[] = [];
    let i = 0;

    while (i < expr.length) {
        const ch = expr.charAt(i);

        if (ch === ' ' || ch === '\t' || ch === '\n' || ch === '\r') {
            i++;
        } else if (ch === '+') {
            tokens.push({ type: TokenType.Plus, value: 0 });
            i++;
        } else if (ch === '-') {
            tokens.push({ type: TokenType.Minus, value: 0 });
            i++;
        } else if (ch === '*') {
            tokens.push({ type: TokenType.Star, value: 0 });
            i++;
        } else if (ch === '/') {
            tokens.push({ type: TokenType.Slash, value: 0 });
            i++;
        } else if (ch === '(') {
            tokens.push({ type: TokenType.LParen, value: 0 });
            i++;
        } else if (ch === ')') {
            tokens.push({ type: TokenType.RParen, value: 0 });
            i++;
        } else if (ch === '.' || isDigit(ch)) {
            const start = i;
            i++;
            while (i < expr.length && (isDigit(expr.charAt(i)) || expr.charAt(i) === '.')) {
                i++;
            }
            const numStr = expr.substring(start, i);
            const val = parseFloat(numStr);
            if (isNaN(val)) {
                throw new Error(`invalid number: ${numStr}`);
            }
            tokens.push({ type: TokenType.Number, value: val });
        } else if (isLetter(ch) || ch === '_') {
            const start = i;
            i++;
            while (i < expr.length && (isLetter(expr.charAt(i)) || isDigit(expr.charAt(i)) || expr.charAt(i) === '_')) {
                i++;
            }
            const name = expr.substring(start, i);
            if (!(name in vars)) {
                throw new Error(`undefined variable: ${name}`);
            }
            tokens.push({ type: TokenType.Number, value: vars[name]! });
        } else {
            throw new Error(`unexpected character: ${ch}`);
        }
    }

    return tokens;
}

function isDigit(ch: string): boolean {
    return ch >= '0' && ch <= '9';
}

function isLetter(ch: string): boolean {
    return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z');
}

class Parser {
    tokens: Token[];
    pos: number;

    constructor(tokens: Token[]) {
        this.tokens = tokens;
        this.pos = 0;
    }

    peek(): Token | null {
        if (this.pos >= this.tokens.length) {
            return null;
        }
        return this.tokens[this.pos]!;
    }

    consume(): void {
        this.pos++;
    }

    parseExpression(): number {
        let left = this.parseTerm();

        while (true) {
            const tok = this.peek();
            if (!tok) break;

            switch (tok.type) {
                case TokenType.Plus:
                    this.consume();
                    left += this.parseTerm();
                    break;
                case TokenType.Minus:
                    this.consume();
                    left -= this.parseTerm();
                    break;
                default:
                    return left;
            }
        }

        return left;
    }

    parseTerm(): number {
        let left = this.parseFactor();

        while (true) {
            const tok = this.peek();
            if (!tok) break;

            switch (tok.type) {
                case TokenType.Star:
                    this.consume();
                    left *= this.parseFactor();
                    break;
                case TokenType.Slash:
                    this.consume();
                    const right = this.parseFactor();
                    if (right === 0) {
                        throw new Error('division by zero');
                    }
                    left /= right;
                    break;
                default:
                    return left;
            }
        }

        return left;
    }

    parseFactor(): number {
        const tok = this.peek();
        if (!tok) {
            throw new Error('unexpected end of expression');
        }

        if (tok.type === TokenType.Minus) {
            this.consume();
            return -this.parseFactor();
        }

        if (tok.type === TokenType.Number) {
            this.consume();
            return tok.value;
        }

        if (tok.type === TokenType.LParen) {
            this.consume();
            const val = this.parseExpression();
            const nextTok = this.peek();
            if (!nextTok || nextTok.type !== TokenType.RParen) {
                throw new Error('missing closing parenthesis');
            }
            this.consume();
            return val;
        }

        throw new Error('unexpected token');
    }
}
