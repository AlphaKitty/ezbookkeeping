export interface ItemField {
    readonly key: string;
    readonly label: string;
    readonly fieldType: 'number' | 'text' | 'enum' | 'date';
    readonly required: boolean;
    readonly editable?: boolean;
    readonly participateInNaming?: boolean;
    readonly options?: string[];
    readonly unit?: string;
    readonly format?: string;
    readonly defaultValue?: string;
    readonly expr?: string;
    readonly sortOrder: number;
}

export interface ItemFieldSchema {
    readonly fields: ItemField[];
}

export interface ItemDefinitionCreateRequest {
    readonly name: string;
    readonly icon: string;
    readonly fieldSchema: ItemFieldSchema;
    readonly expensePricingExpr: string;
    readonly incomePricingExpr: string;
    readonly incomeCategoryId: string;
    readonly expenseCategoryId: string;
}

export interface ItemDefinitionModifyRequest {
    readonly id: string;
    readonly name: string;
    readonly icon: string;
    readonly fieldSchema: ItemFieldSchema;
    readonly expensePricingExpr: string;
    readonly incomePricingExpr: string;
    readonly incomeCategoryId: string;
    readonly expenseCategoryId: string;
}

export interface ItemDefinitionDeleteRequest {
    readonly id: string;
}

export interface ItemDefinitionInfoResponse {
    readonly id: string;
    readonly name: string;
    readonly icon: string;
    readonly fieldSchema: ItemFieldSchema | null;
    readonly expensePricingExpr: string;
    readonly incomePricingExpr: string;
    readonly incomeCategoryId: string;
    readonly expenseCategoryId: string;
}

export enum ItemFieldType {
    NUMBER = 'number',
    TEXT = 'text',
    ENUM = 'enum',
    DATE = 'date',
}

export const ITEM_FIELD_TYPE_OPTIONS = [
    { value: ItemFieldType.NUMBER, label: '数字' },
    { value: ItemFieldType.TEXT, label: '文本' },
    { value: ItemFieldType.ENUM, label: '枚举/下拉' },
    { value: ItemFieldType.DATE, label: '时间' },
];

export const ITEM_DATETIME_FORMAT_OPTIONS = [
    { value: 'YYYY-MM-DD', label: '年月日' },
    { value: 'YYYY-MM-DD HH:mm:ss', label: '年月日时分秒' },
];
