<template>
    <v-row class="match-height">
        <v-col cols="12">
            <v-card>
                <template #title>
                    <div class="title-and-toolbar d-flex align-center">
                        <span>{{ tt('Inventory Records') }}</span>
                        <v-btn class="ms-3" color="default" variant="outlined" :disabled="loading" @click="openCreateDialog">{{ tt('Add') }}</v-btn>
                        <v-btn density="compact" color="default" variant="text" size="24" class="ms-2" :icon="true" :disabled="loading" :loading="loading" @click="reload">
                            <template #loader><v-progress-circular indeterminate size="20"/></template>
                            <v-icon :icon="mdiRefresh" size="24"/>
                            <v-tooltip activator="parent">{{ tt('Refresh') }}</v-tooltip>
                        </v-btn>
                        <v-spacer/>
                    </div>
                </template>

                <!-- Item type tabs -->
                <div class="d-flex ga-2 px-4 pt-2 flex-wrap">
                    <v-btn density="compact" :color="!activeItemDefId ? 'primary' : 'default'" :variant="!activeItemDefId ? 'tonal' : 'text'" size="small" @click="activeItemDefId = ''">{{ tt('All') }}</v-btn>
                    <v-btn v-for="def in itemDefinitions" :key="def.id" density="compact" :color="activeItemDefId === def.id ? 'primary' : 'default'" :variant="activeItemDefId === def.id ? 'tonal' : 'text'" size="small" @click="activeItemDefId = def.id">{{ def.name }}</v-btn>
                </div>

                <v-data-table :headers="displayHeaders" :items="displayRecords" :loading="loading" :no-data-text="tt('No data')" items-per-page="-1" hide-default-footer>
                    <template #item.itemDefinitionName="{ item }">
                        <v-chip density="compact" variant="tonal" size="small">{{ item.itemDefinitionName }}</v-chip>
                    </template>
                    <template #item.status="{ item }">
                        <v-chip density="compact" :color="statusColor(item.status)" size="small">{{ statusLabel(item.status) }}</v-chip>
                    </template>
                    <template #item.updatedTime="{ item }">
                        {{ formatTime(item.updatedUnixTime) }}
                    </template>
                    <!-- Dynamic field slots -->
                    <template v-for="field in activeViewFields" :key="field.key" #[`item.${field.key}`]="{ item }">
                        {{ item[field.key] ?? '--' }}
                    </template>
                    <template #item.actions="{ item }">
                        <v-btn density="compact" variant="text" :icon="true" size="small" @click="openEditDialog(item)">
                            <v-icon :icon="mdiPencilOutline" size="18"/>
                        </v-btn>
                        <v-btn density="compact" variant="text" :icon="true" size="small" color="error" @click="confirmDelete(item)">
                            <v-icon :icon="mdiDeleteOutline" size="18"/>
                        </v-btn>
                    </template>
                </v-data-table>
            </v-card>
        </v-col>
    </v-row>

    <!-- Create/Edit Dialog -->
    <v-dialog v-model="showDialog" width="640" persistent>
        <v-card class="pa-sm-1 pa-md-2">
            <template #title>
                <h4 class="text-h4">{{ isEditing ? tt('Edit Inventory Record') : tt('Add Inventory Record') }}</h4>
            </template>
            <v-card-text>
                <v-select v-model="form.itemDefinitionId" :label="tt('Item Type')" :items="itemDefOptions" item-title="name" item-value="id" density="compact" variant="outlined" class="mb-4" @update:model-value="onItemTypeChange"/>

                <template v-if="currentItemDefinition?.fieldSchema?.fields?.length">
                    <v-divider class="mb-3"/>
                    <div class="text-subtitle-2 mb-3">{{ currentItemDefinition.name }}</div>
                    <v-row>
                        <v-col v-for="field in currentItemDefinition.fieldSchema.fields" :key="field.key" cols="12" :md="field.fieldType === 'text' ? 12 : 6">
                            <v-text-field v-if="field.fieldType === 'number'"
                                v-model.number="fieldValues[field.key]"
                                :label="field.key"
                                :suffix="field.unit"
                                type="number"
                                density="compact" variant="outlined"
                                :rules="field.required ? [required] : []"/>
                            <v-text-field v-else-if="field.fieldType === 'text'"
                                v-model="fieldValues[field.key]"
                                :label="field.key"
                                density="compact" variant="outlined"
                                :rules="field.required ? [required] : []"/>
                            <v-select v-else-if="field.fieldType === 'enum'"
                                v-model="fieldValues[field.key]"
                                :label="field.key"
                                :items="(field as any).options || []"
                                density="compact" variant="outlined"
                                :rules="field.required ? [required] : []"/>
                            <v-text-field v-else-if="field.fieldType === 'date'"
                                v-model="fieldValues[field.key]"
                                :label="field.key"
                                :type="(field as any).format === 'YYYY-MM-DD HH:mm:ss' ? 'datetime-local' : 'date'"
                                density="compact" variant="outlined"
                                :rules="field.required ? [required] : []"/>
                        </v-col>
                    </v-row>
                </template>

                <p v-if="!currentItemDefinition?.fieldSchema?.fields?.length && form.itemDefinitionId" class="text-caption text-disabled mt-4">{{ tt('This item type has no custom fields defined') }}</p>

                <v-alert v-if="formError" type="error" variant="tonal" density="compact" class="mt-4" closable @click:close="formError = ''">{{ formError }}</v-alert>
            </v-card-text>
            <v-card-actions>
                <v-spacer/>
                <v-btn variant="text" @click="showDialog = false">{{ tt('Cancel') }}</v-btn>
                <v-btn color="primary" variant="tonal" :loading="saving" @click="save">{{ tt('Save') }}</v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>

    <!-- Delete Confirm Dialog -->
    <v-dialog v-model="showDeleteDialog" :max-width="400" persistent>
        <v-card>
            <template #title>{{ tt('Delete Inventory Record') }}</template>
            <v-card-text>{{ tt('Are you sure you want to delete this inventory record?') }}</v-card-text>
            <v-card-actions>
                <v-spacer/>
                <v-btn variant="text" @click="showDeleteDialog = false">{{ tt('Cancel') }}</v-btn>
                <v-btn color="error" variant="tonal" :loading="deleting" @click="doDelete">{{ tt('Delete') }}</v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useI18n } from '@/locales/helpers.ts';
import { mdiRefresh, mdiPencilOutline, mdiDeleteOutline } from '@mdi/js';

import api from '@/lib/services.ts';
import type { InventoryRecordInfoResponse } from '@/models/inventory_record.ts';
import { INVENTORY_STATUS_OPTIONS } from '@/models/inventory_record.ts';
import type { ItemDefinitionInfoResponse } from '@/models/item_definition.ts';

type DisplayRecord = InventoryRecordInfoResponse & Record<string, any>;

const { tt } = useI18n();
function required(v: any): true | string {
    if (v === null || v === undefined || v === '') return tt('Required');
    if (typeof v === 'number' && isNaN(v)) return tt('Required');
    return true;
}

const loading = ref(false);
const saving = ref(false);
const deleting = ref(false);
const records = ref<InventoryRecordInfoResponse[]>([]);
const itemDefinitions = ref<ItemDefinitionInfoResponse[]>([]);
const showDialog = ref(false);
const showDeleteDialog = ref(false);
const isEditing = ref(false);
const deleteTarget = ref<InventoryRecordInfoResponse | null>(null);
const editingId = ref<string>('');
const activeItemDefId = ref<string>('');

const currentItemDefinition = ref<ItemDefinitionInfoResponse | null>(null);
const fieldValues = ref<Record<string, any>>({});
const formError = ref('');

const form = ref({
    itemDefinitionId: '',
});

const itemDefOptions = computed(() =>
    itemDefinitions.value.map(d => ({ id: d.id, name: d.name }))
);

const activeViewDef = computed(() =>
    activeItemDefId.value
        ? itemDefinitions.value.find(d => d.id === activeItemDefId.value) || null
        : null
);

const activeViewFields = computed(() =>
    activeViewDef.value?.fieldSchema?.fields || []
);

const displayHeaders = computed(() => {
    const h: any[] = [
        { title: tt('Item Type'), key: 'itemDefinitionName', align: 'center' as const },
    ];

    if (activeViewDef.value) {
        for (const field of activeViewFields.value) {
            h.push({ title: field.key, key: field.key, align: 'center' as const });
        }
    }

    h.push({ title: tt('Status'), key: 'status', align: 'center' as const });

    if (!activeItemDefId.value) {
        h.push({ title: tt('Updated Time'), key: 'updatedTime', align: 'center' as const });
    }

    h.push({ title: tt('Actions'), key: 'actions', sortable: false, align: 'center' as const });
    return h;
});

const displayRecords = computed<DisplayRecord[]>(() => {
    let result = records.value;

    if (activeItemDefId.value) {
        result = result.filter(r => r.itemDefinitionId === activeItemDefId.value);
    }

    return result.map(r => ({
        ...r,
        ...(r.fieldValues?.values || {}),
        updatedTime: r.updatedUnixTime,
    }));
});

function statusColor(status: string): string {
    switch (status) {
        case 'in_stock': return 'success';
        case 'reserved': return 'warning';
        case 'sold_out': return 'error';
        default: return 'default';
    }
}

function statusLabel(status: string): string {
    const found = INVENTORY_STATUS_OPTIONS.find(o => o.value === status);
    return found?.label || status;
}

function formatTime(unixTime: number): string {
    if (!unixTime) return '--';
    return new Date(unixTime * 1000).toLocaleDateString();
}

function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape' && showDialog.value) {
        showDialog.value = false;
    }
}

async function reload() {
    loading.value = true;
    try {
        const [recordsResp, defsResp] = await Promise.all([
            api.getInventoryRecords(),
            api.getItemDefinitions(),
        ]);
        records.value = recordsResp.data.result;
        itemDefinitions.value = defsResp.data.result;
    } finally {
        loading.value = false;
    }
}

async function loadItemDefinition(itemDefId: string, existingValues?: Record<string, any>) {
    if (!itemDefId) {
        currentItemDefinition.value = null;
        fieldValues.value = {};
        return;
    }
    const resp = await api.getItemDefinition({ id: itemDefId });
    currentItemDefinition.value = resp.data.result;
    fieldValues.value = existingValues || {};
}

async function onItemTypeChange(newId: string) {
    formError.value = '';
    await loadItemDefinition(newId);
}

function openCreateDialog() {
    isEditing.value = false;
    editingId.value = '';
    form.value = { itemDefinitionId: '' };
    currentItemDefinition.value = null;
    fieldValues.value = {};
    formError.value = '';
    showDialog.value = true;
}

async function openEditDialog(item: InventoryRecordInfoResponse) {
    isEditing.value = true;
    editingId.value = item.id;
    form.value = { itemDefinitionId: item.itemDefinitionId };
    formError.value = '';
    await loadItemDefinition(item.itemDefinitionId, item.fieldValues?.values);
    showDialog.value = true;
}

function validateRequiredFields(): string | null {
    if (!currentItemDefinition.value?.fieldSchema?.fields?.length) return null;
    for (const field of currentItemDefinition.value.fieldSchema.fields) {
        if (!field.required) continue;
        const v = fieldValues.value[field.key];
        if (v === null || v === undefined || v === '') return `${tt('Required')}: ${field.key}`;
        if (typeof v === 'number' && isNaN(v)) return `${tt('Required')}: ${field.key}`;
    }
    return null;
}

async function save() {
    const err = validateRequiredFields();
    if (err) {
        formError.value = err;
        return;
    }

    saving.value = true;
    try {
        const fieldValuesPayload = currentItemDefinition.value?.fieldSchema?.fields?.length
            ? { values: { ...fieldValues.value } }
            : null;

        if (isEditing.value) {
            await api.modifyInventoryRecord({
                id: editingId.value,
                itemDefinitionId: form.value.itemDefinitionId,
                warehouseId: '0',
                fieldValues: fieldValuesPayload,
                quantity: 0,
                unit: '',
                unitPrice: 0,
                transporter: '',
                batchNo: '',
                status: 'in_stock' as any,
                comment: '',
            });
        } else {
            await api.addInventoryRecord({
                itemDefinitionId: form.value.itemDefinitionId,
                warehouseId: '0',
                fieldValues: fieldValuesPayload,
                quantity: 0,
                unit: '',
                unitPrice: 0,
                transporter: '',
                batchNo: '',
                comment: '',
            });
        }
        showDialog.value = false;
        await reload();
    } finally {
        saving.value = false;
    }
}

function confirmDelete(item: InventoryRecordInfoResponse) {
    deleteTarget.value = item;
    showDeleteDialog.value = true;
}

async function doDelete() {
    if (!deleteTarget.value) return;
    deleting.value = true;
    try {
        await api.deleteInventoryRecord({ id: deleteTarget.value.id });
        showDeleteDialog.value = false;
        deleteTarget.value = null;
        await reload();
    } finally {
        deleting.value = false;
    }
}

onMounted(() => {
    window.addEventListener('keydown', handleKeydown);
    reload();
});

onUnmounted(() => {
    window.removeEventListener('keydown', handleKeydown);
});
</script>
