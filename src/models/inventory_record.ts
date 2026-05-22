export interface ItemFieldValues {
    readonly values: Record<string, unknown>;
}

export type InventoryStatus = 'in_stock' | 'reserved' | 'sold_out';
export type InventoryAction = 'none' | 'stock_in' | 'stock_out';

export interface InventoryRecordCreateRequest {
    readonly itemDefinitionId: string;
    readonly warehouseId: string;
    readonly fieldValues: ItemFieldValues | null;
    readonly quantity: number;
    readonly unit: string;
    readonly unitPrice: number;
    readonly transporter: string;
    readonly batchNo: string;
    readonly comment: string;
}

export interface InventoryRecordModifyRequest {
    readonly id: string;
    readonly itemDefinitionId: string;
    readonly warehouseId: string;
    readonly fieldValues: ItemFieldValues | null;
    readonly quantity: number;
    readonly unit: string;
    readonly unitPrice: number;
    readonly transporter: string;
    readonly batchNo: string;
    readonly status: InventoryStatus;
    readonly comment: string;
}

export interface InventoryRecordDeleteRequest {
    readonly id: string;
}

export interface InventoryRecordInfoResponse {
    readonly id: string;
    readonly itemDefinitionId: string;
    readonly itemDefinitionName: string;
    readonly warehouseId: string;
    readonly fieldValues: ItemFieldValues | null;
    readonly quantity: number;
    readonly unit: string;
    readonly unitPrice: number;
    readonly transporter: string;
    readonly batchNo: string;
    readonly status: InventoryStatus;
    readonly comment: string;
    readonly createdUnixTime: number;
    readonly updatedUnixTime: number;
}

export const INVENTORY_STATUS_OPTIONS = [
    { value: 'in_stock', label: '在库' },
    { value: 'reserved', label: '已预留' },
    { value: 'sold_out', label: '已售出' },
] as const;

export const INVENTORY_ACTION_OPTIONS = [
    { value: 'none', label: '无' },
    { value: 'stock_in', label: '入库' },
    { value: 'stock_out', label: '出库' },
] as const;
