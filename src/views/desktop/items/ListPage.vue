<template>
    <v-row class="match-height">
        <v-col cols="12">
            <v-card>
                <template #title>
                    <div class="title-and-toolbar d-flex align-center">
                        <span>{{ tt('Item Definitions') }}</span>
                        <v-btn class="ms-3" color="default" variant="outlined" :disabled="loading" @click="openCreateDialog">{{ tt('Add') }}</v-btn>
                        <v-btn density="compact" color="default" variant="text" size="24" class="ms-2" :icon="true" :disabled="loading" :loading="loading" @click="reload">
                            <template #loader><v-progress-circular indeterminate size="20"/></template>
                            <v-icon :icon="mdiRefresh" size="24"/>
                            <v-tooltip activator="parent">{{ tt('Refresh') }}</v-tooltip>
                        </v-btn>
                        <v-spacer/>
                    </div>
                </template>
                <v-data-table :headers="headers" :items="definitions" :loading="loading" :no-data-text="tt('No data')" items-per-page="-1" hide-default-footer>
                    <template #item.name="{ item }">
                        <div class="d-flex align-center justify-center">
                            <ItemIcon icon-type="category" :icon-id="item.icon" class="me-2"/>
                            {{ item.name }}
                        </div>
                    </template>
                    <template #item.fieldCount="{ item }">
                        {{ item.fieldSchema?.fields?.length || 0 }}
                    </template>
                    <template #item.pricingExpr="{ item }">
                        <code v-if="item.pricingExpr" class="text-caption">{{ item.pricingExpr }}</code>
                        <span v-else class="text-caption text-disabled">--</span>
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
    <v-dialog v-model="showDialog" width="800" persistent>
        <v-card class="pa-sm-1 pa-md-2">
            <template #title>
                <h4 class="text-h4">{{ isEditing ? tt('Edit Item Definition') : tt('Add Item Definition') }}</h4>
            </template>
            <v-card-text>
                <v-row>
                    <v-col cols="12" md="6">
                        <v-text-field v-model="form.name" :label="tt('Name')" density="compact" variant="outlined" :rules="[required]"/>
                    </v-col>
                    <v-col cols="12" md="6">
                        <icon-select icon-type="category"
                                     :all-icon-infos="ALL_CATEGORY_ICONS"
                                     :label="tt('Item Icon')"
                                     :color="''"
                                     :disabled="saving"
                                     v-model="form.icon" />
                    </v-col>
                </v-row>

                <div class="text-subtitle-2 mb-2 mt-4">{{ tt('Pricing Expression') }}</div>
                <v-text-field v-model="form.pricingExpr" :placeholder="tt('Pricing Expression (e.g. weight * unit_price)')" density="compact" variant="outlined" class="mb-2"/>
                <div class="mb-4">
                    <span class="text-caption text-disabled me-2" v-if="validFieldKeys.length">{{ tt('Select field to insert into expression') }}:</span>
                    <v-chip v-for="key in validFieldKeys" :key="key" density="compact" size="small" class="me-1 mb-1" @click="insertFieldKey(key)">
                        {{ key }}
                    </v-chip>
                    <span v-if="!validFieldKeys.length" class="text-caption text-disabled">{{ tt('Add fields below to use in pricing expression') }}</span>
                </div>

                <v-divider class="mb-4"/>

                <div class="text-subtitle-2 mb-2">{{ tt('Transaction Categories') }}</div>
                <v-row>
                    <v-col cols="12" md="6">
                        <two-column-select primary-key-field="id" primary-value-field="id" primary-title-field="name"
                                           primary-icon-field="icon" primary-icon-type="category" primary-color-field="color"
                                           primary-hidden-field="hidden" primary-sub-items-field="subCategories"
                                           secondary-key-field="id" secondary-value-field="id" secondary-title-field="name"
                                           secondary-icon-field="icon" secondary-icon-type="category" secondary-color-field="color"
                                           secondary-hidden-field="hidden"
                                           :disabled="saving"
                                           :enable-filter="true" :filter-placeholder="tt('Find category')" :filter-no-items-text="tt('No available category')"
                                           :show-selection-primary-text="true"
                                           :custom-selection-primary-text="getCategoryPrimaryName(form.incomeCategoryId, incomeCategories)"
                                           :custom-selection-secondary-text="getCategorySecondaryName(form.incomeCategoryId, incomeCategories)"
                                           :label="tt('Income Category')" :placeholder="tt('Income Category')"
                                           :items="incomeCategories"
                                           v-model="form.incomeCategoryId">
                        </two-column-select>
                    </v-col>
                    <v-col cols="12" md="6">
                        <two-column-select primary-key-field="id" primary-value-field="id" primary-title-field="name"
                                           primary-icon-field="icon" primary-icon-type="category" primary-color-field="color"
                                           primary-hidden-field="hidden" primary-sub-items-field="subCategories"
                                           secondary-key-field="id" secondary-value-field="id" secondary-title-field="name"
                                           secondary-icon-field="icon" secondary-icon-type="category" secondary-color-field="color"
                                           secondary-hidden-field="hidden"
                                           :disabled="saving"
                                           :enable-filter="true" :filter-placeholder="tt('Find category')" :filter-no-items-text="tt('No available category')"
                                           :show-selection-primary-text="true"
                                           :custom-selection-primary-text="getCategoryPrimaryName(form.expenseCategoryId, expenseCategories)"
                                           :custom-selection-secondary-text="getCategorySecondaryName(form.expenseCategoryId, expenseCategories)"
                                           :label="tt('Expense Category')" :placeholder="tt('Expense Category')"
                                           :items="expenseCategories"
                                           v-model="form.expenseCategoryId">
                        </two-column-select>
                    </v-col>
                </v-row>

                <v-divider class="mb-4"/>

                <div class="d-flex align-center mb-3">
                    <span class="text-subtitle-2">{{ tt('Fields') }}</span>
                    <v-spacer/>
                    <v-btn density="compact" variant="outlined" size="small" @click="addField" :prepend-icon="mdiPlus">{{ tt('Add Field') }}</v-btn>
                </div>

                <v-card variant="outlined" class="mb-3 pa-3" v-for="(field, idx) in form.fields" :key="idx">
                    <div class="d-flex justify-end mb-2">
                        <v-btn density="compact" variant="text" :icon="true" size="small" color="error" @click="removeField(idx)">
                            <v-icon :icon="mdiClose" size="16"/>
                        </v-btn>
                    </div>
                    <v-row dense>
                        <v-col cols="12" :sm="field.fieldType === 'number' || field.fieldType === 'date' ? 5 : 7">
                            <v-text-field v-model="field.key" :label="tt('Key')" density="compact" variant="outlined"/>
                        </v-col>
                        <v-col cols="12" :sm="field.fieldType === 'number' || field.fieldType === 'date' ? 3 : 5">
                            <v-select v-model="field.fieldType" :label="tt('Field Type')" :items="fieldTypeOptions" item-title="label" item-value="value" density="compact" variant="outlined"/>
                        </v-col>
                        <v-col v-if="field.fieldType === 'number'" cols="12" sm="4">
                            <v-text-field v-model="field.unit" :label="tt('Unit')" density="compact" variant="outlined"/>
                        </v-col>
                        <v-col v-if="field.fieldType === 'date'" cols="12" sm="4">
                            <v-select v-model="field.format" :label="tt('Format')" :items="dateTimeFormatOptions" item-title="label" item-value="value" density="compact" variant="outlined"/>
                        </v-col>
                    </v-row>
                    <v-row dense class="mt-1">
                        <v-col cols="12" sm="4" class="d-flex align-center">
                            <v-checkbox v-model="field.required" :label="tt('Required')" density="compact" hide-details/>
                        </v-col>
                        <v-col cols="12" sm="4" class="d-flex align-center">
                            <v-checkbox v-model="field.editable" :label="tt('Editable')" density="compact" hide-details/>
                        </v-col>
                        <v-col cols="12" sm="4" class="d-flex align-center">
                            <v-checkbox v-model="field.participateInNaming" :label="tt('Participate in Naming')" density="compact" hide-details/>
                        </v-col>
                    </v-row>
                    <div v-if="field.fieldType === 'enum'">
                        <div class="d-flex align-center mb-1 mt-2">
                            <span class="text-caption">{{ tt('Options') }}</span>
                            <v-spacer/>
                            <v-btn density="compact" variant="text" size="small" @click="addOption(idx)" :prepend-icon="mdiPlus">{{ tt('Add Option') }}</v-btn>
                        </div>
                        <v-row dense>
                            <v-col cols="12" sm="6" v-for="(_opt, optIdx) in (field.options || [])" :key="optIdx" class="d-flex align-center ga-2">
                                <v-text-field v-model="field.options![optIdx]" density="compact" variant="outlined" hide-details/>
                                <v-btn density="compact" variant="text" :icon="true" size="small" color="error" @click="removeOption(idx, optIdx)">
                                    <v-icon :icon="mdiClose" size="14"/>
                                </v-btn>
                            </v-col>
                        </v-row>
                    </div>
                </v-card>
            </v-card-text>
            <v-alert v-if="validationError" type="error" variant="tonal" density="compact" class="mx-4 mb-2" closable @click:close="validationError = ''">{{ validationError }}</v-alert>
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
            <template #title>{{ tt('Delete Item Definition') }}</template>
            <v-card-text>{{ tt('Are you sure you want to delete this item?') }}</v-card-text>
            <v-card-actions>
                <v-spacer/>
                <v-btn variant="text" @click="showDeleteDialog = false">{{ tt('Cancel') }}</v-btn>
                <v-btn color="error" variant="tonal" :loading="deleting" @click="doDelete">{{ tt('Delete') }}</v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useI18n } from '@/locales/helpers.ts';
import { mdiRefresh, mdiPencilOutline, mdiDeleteOutline, mdiPlus, mdiClose } from '@mdi/js';

import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { CategoryType } from '@/core/category.ts';
import {
    getTransactionPrimaryCategoryName,
    getTransactionSecondaryCategoryName
} from '@/lib/category.ts';

import api from '@/lib/services.ts';
import type { ItemDefinitionInfoResponse } from '@/models/item_definition.ts';
import type { TransactionCategory } from '@/models/transaction_category.ts';
import { ITEM_FIELD_TYPE_OPTIONS, ITEM_DATETIME_FORMAT_OPTIONS, ItemFieldType } from '@/models/item_definition.ts';

import IconSelect from '@/components/desktop/IconSelect.vue';
import ItemIcon from '@/components/desktop/ItemIcon.vue';
import { ALL_CATEGORY_ICONS } from '@/consts/icon.ts';

interface MutableItemField {
    key: string;
    label: string;
    fieldType: 'number' | 'text' | 'enum' | 'date';
    required: boolean;
    editable: boolean;
    participateInNaming: boolean;
    options?: string[];
    unit?: string;
    format?: string;
    defaultValue?: string;
    sortOrder: number;
}

const { tt } = useI18n();

const transactionCategoriesStore = useTransactionCategoriesStore();
const incomeCategories = computed<TransactionCategory[]>(() =>
    transactionCategoriesStore.allTransactionCategories[CategoryType.Income] || []
);
const expenseCategories = computed<TransactionCategory[]>(() =>
    transactionCategoriesStore.allTransactionCategories[CategoryType.Expense] || []
);

function getCategoryPrimaryName(categoryId: string, categories: TransactionCategory[]): string {
    return getTransactionPrimaryCategoryName(categoryId, categories);
}
function getCategorySecondaryName(categoryId: string, categories: TransactionCategory[]): string {
    return getTransactionSecondaryCategoryName(categoryId, categories);
}

const required = (v: string) => !!v || tt('Required');

const loading = ref(false);
const saving = ref(false);
const deleting = ref(false);
const validationError = ref('');
const definitions = ref<ItemDefinitionInfoResponse[]>([]);
const showDialog = ref(false);
const showDeleteDialog = ref(false);
const isEditing = ref(false);
const deleteTarget = ref<ItemDefinitionInfoResponse | null>(null);
const editingId = ref<string>('');

const fieldTypeOptions = computed(() => ITEM_FIELD_TYPE_OPTIONS.map(opt => ({
    value: opt.value,
    label: tt(opt.value === ItemFieldType.NUMBER ? 'Number' : opt.value === ItemFieldType.TEXT ? 'Text' : opt.value === ItemFieldType.ENUM ? 'Enum' : 'Date'),
})));

const dateTimeFormatOptions = computed(() => ITEM_DATETIME_FORMAT_OPTIONS.map(opt => ({
    value: opt.value,
    label: tt(opt.value),
})));

const emptyField = (): MutableItemField => ({
    key: '', label: '', fieldType: 'number', required: true, editable: false, participateInNaming: false, unit: '', sortOrder: 0,
});

const form = ref({
    name: '',
    icon: '',
    pricingExpr: '',
    incomeCategoryId: '',
    expenseCategoryId: '',
    fields: [] as MutableItemField[],
});

const validFieldKeys = computed(() =>
    form.value.fields.filter(f => f.key).map(f => f.key)
);

const headers = [
    { title: tt('Name'), key: 'name', align: 'center' as const },
    { title: tt('Fields'), key: 'fieldCount', align: 'center' as const },
    { title: tt('Pricing Expression'), key: 'pricingExpr', align: 'center' as const },
    { title: tt('Actions'), key: 'actions', sortable: false, align: 'center' as const },
];

async function reload() {
    loading.value = true;
    try {
        const resp = await api.getItemDefinitions();
        definitions.value = resp.data.result;
    } finally {
        loading.value = false;
    }
}

function openCreateDialog() {
    isEditing.value = false;
    editingId.value = '';
    form.value = { name: '', icon: '', pricingExpr: '', incomeCategoryId: '', expenseCategoryId: '', fields: [] };
    showDialog.value = true;
}

function openEditDialog(item: ItemDefinitionInfoResponse) {
    isEditing.value = true;
    editingId.value = item.id;
    form.value = {
        name: item.name,
        icon: item.icon,
        pricingExpr: item.pricingExpr,
        incomeCategoryId: item.incomeCategoryId || '',
        expenseCategoryId: item.expenseCategoryId || '',
        fields: item.fieldSchema?.fields
            ? JSON.parse(JSON.stringify(item.fieldSchema.fields))
            : [],
    };
    showDialog.value = true;
}

function addField() {
    form.value.fields.push(emptyField());
}

function removeField(idx: number) {
    form.value.fields.splice(idx, 1);
}

function addOption(fieldIdx: number) {
    const field = form.value.fields[fieldIdx]!;
    if (!field.options) {
        field.options = [];
    }
    field.options.push('');
}

function removeOption(fieldIdx: number, optIdx: number) {
    const field = form.value.fields[fieldIdx]!;
    if (field.options) {
        field.options.splice(optIdx, 1);
    }
}

function insertFieldKey(key: string) {
    form.value.pricingExpr += (form.value.pricingExpr ? ' ' : '') + key;
}

async function save() {
    if (!form.value.incomeCategoryId) {
        validationError.value = tt('Income Category') + ': ' + tt('Required');
        return;
    }
    if (!form.value.expenseCategoryId) {
        validationError.value = tt('Expense Category') + ': ' + tt('Required');
        return;
    }
    validationError.value = '';
    saving.value = true;
    try {
        const fieldSchema = { fields: form.value.fields.map((f, i) => ({ ...f, sortOrder: i })) };
        if (isEditing.value) {
            await api.modifyItemDefinition({
                id: editingId.value,
                name: form.value.name,
                icon: form.value.icon,
                fieldSchema,
                pricingExpr: form.value.pricingExpr,
                incomeCategoryId: form.value.incomeCategoryId || '0',
                expenseCategoryId: form.value.expenseCategoryId || '0',
            });
        } else {
            await api.addItemDefinition({
                name: form.value.name,
                icon: form.value.icon,
                fieldSchema,
                pricingExpr: form.value.pricingExpr,
                incomeCategoryId: form.value.incomeCategoryId || '0',
                expenseCategoryId: form.value.expenseCategoryId || '0',
            });
        }
        showDialog.value = false;
        await reload();
    } finally {
        saving.value = false;
    }
}

function confirmDelete(item: ItemDefinitionInfoResponse) {
    deleteTarget.value = item;
    showDeleteDialog.value = true;
}

async function doDelete() {
    if (!deleteTarget.value) return;
    deleting.value = true;
    try {
        await api.deleteItemDefinition({ id: deleteTarget.value.id });
        showDeleteDialog.value = false;
        deleteTarget.value = null;
        await reload();
    } finally {
        deleting.value = false;
    }
}

onMounted(() => {
    reload();
    transactionCategoriesStore.loadAllCategories({ force: false });
});
</script>
