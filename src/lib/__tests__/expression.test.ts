import { describe, expect, it } from 'vitest';

import { evaluateExpression, evaluateFieldExpressions, type FieldExpr } from '@/lib/expression.ts';

describe('evaluateFieldExpressions', () => {
    it('should compute area from length * width (user scenario)', () => {
        const fields: FieldExpr[] = [
            { key: 'area', expr: 'length * width' },
        ];
        const values = { length: 3, width: 2 };
        const result = evaluateFieldExpressions(fields, values);
        expect(result).toEqual({ area: 6 });
    });

    it('should return empty when dependency missing', () => {
        const fields: FieldExpr[] = [
            { key: 'area', expr: 'length * width' },
        ];
        const values = { length: 3 };
        const result = evaluateFieldExpressions(fields, values);
        expect(result).toEqual({});
    });

    it('should compute chain dependency', () => {
        const fields: FieldExpr[] = [
            { key: 'subtotal', expr: 'price * qty' },
            { key: 'tax', expr: 'subtotal * 0.13' },
        ];
        const values = { price: 100, qty: 2 };
        const result = evaluateFieldExpressions(fields, values);
        expect(result).toEqual({ subtotal: 200, tax: 26 });
    });

    it('should handle fields with only constants', () => {
        const fields: FieldExpr[] = [
            { key: 'doubled', expr: 'x * 2' },
        ];
        const values = { x: 5 };
        const result = evaluateFieldExpressions(fields, values);
        expect(result).toEqual({ doubled: 10 });
    });
});

describe('evaluateExpression', () => {
    it('should return zero for empty expression', () => {
        expect(evaluateExpression('', {})).toBe(0);
    });

    it('should evaluate a plain number', () => {
        expect(evaluateExpression('42', {})).toBe(42);
    });

    it('should evaluate simple multiplication with variables', () => {
        expect(evaluateExpression('price * qty', { price: 10, qty: 5 })).toBe(50);
    });

    it('should evaluate addition', () => {
        expect(evaluateExpression('a + b', { a: 100, b: 200 })).toBe(300);
    });

    it('should evaluate subtraction', () => {
        expect(evaluateExpression('a - b', { a: 100, b: 30 })).toBe(70);
    });

    it('should evaluate division', () => {
        expect(evaluateExpression('a / b', { a: 100, b: 4 })).toBe(25);
    });

    it('should respect operator precedence (multiplication before addition)', () => {
        expect(evaluateExpression('2 + 3 * 4', {})).toBe(14);
    });

    it('should handle parentheses', () => {
        expect(evaluateExpression('(2 + 3) * 4', {})).toBe(20);
    });

    it('should handle unary minus', () => {
        expect(evaluateExpression('-price', { price: 10 })).toBe(-10);
    });

    it('should handle decimal numbers', () => {
        expect(evaluateExpression('2.5 * 4', {})).toBe(10);
    });

    it('should throw on undefined variable', () => {
        expect(() => evaluateExpression('undefined_var + 1', {})).toThrow();
    });

    it('should throw on division by zero', () => {
        expect(() => evaluateExpression('x / 0', { x: 1 })).toThrow();
    });
});
