<template>
    <f7-page>
        <f7-navbar>
            <f7-nav-left :back-link="tt('Back')" :class="{ 'disabled': saving }"></f7-nav-left>
            <f7-nav-title>{{ isEditing ? tt('Edit Inventory Record') : tt('Add Inventory Record') }}</f7-nav-title>
            <f7-nav-right :class="{ 'disabled': saving }">
                <f7-link icon-f7="checkmark_alt" @click="save" :class="{ 'disabled': saving }"></f7-link>
            </f7-nav-right>
        </f7-navbar>

        <f7-block-title class="margin-top">{{ tt('Basic Information') }}</f7-block-title>
        <f7-list strong inset>
            <f7-list-item :title="tt('Item Type')"
                          :after="selectedItemDefName || ''"
                          link="#" @click="showItemTypeSheet = true" />
        </f7-list>

        <template v-if="currentItemDefinition?.fieldSchema?.fields?.length">
            <f7-block-title>{{ currentItemDefinition.name }}</f7-block-title>
            <f7-list strong inset>
                <template v-for="field in currentItemDefinition.fieldSchema.fields" :key="field.key">
                    <f7-list-input v-if="field.fieldType === 'number'"
                                   type="number"
                                   :label="field.key"
                                   :placeholder="field.key"
                                   :required="field.required"
                                   :error-message="field.required ? tt('Required') : ''"
                                   :error-message-force="field.required && !isFieldValueSet(field.key)"
                                   :suffix="field.unit"
                                   clear-button
                                   v-model:value="fieldValues[field.key]" />

                    <f7-list-input v-else-if="field.fieldType === 'text'"
                                   type="text"
                                   :label="field.key"
                                   :placeholder="field.key"
                                   :required="field.required"
                                   :error-message="field.required ? tt('Required') : ''"
                                   :error-message-force="field.required && !isFieldValueSet(field.key)"
                                   clear-button
                                   v-model:value="fieldValues[field.key]" />

                    <f7-list-item v-else-if="field.fieldType === 'enum'"
                                  :title="field.key"
                                  :required="field.required"
                                  :after="fieldValues[field.key] || tt('Optional')"
                                  link="#" @click="openEnumSheet(field)" />

                    <f7-list-input v-else-if="field.fieldType === 'date'"
                                   :type="field.format === 'YYYY-MM-DD HH:mm:ss' ? 'datetime-local' : 'date'"
                                   :label="field.key"
                                   :placeholder="field.key"
                                   :required="field.required"
                                   :error-message="field.required ? tt('Required') : ''"
                                   :error-message-force="field.required && !isFieldValueSet(field.key)"
                                   clear-button
                                   v-model:value="fieldValues[field.key]" />
                </template>
            </f7-list>
        </template>

        <f7-list strong inset v-if="form.itemDefinitionId && !currentItemDefinition?.fieldSchema?.fields?.length">
            <f7-list-item :title="tt('This item type has no custom fields defined')" />
        </f7-list>

        <!-- Item Type Sheet -->
        <f7-actions close-by-outside-click close-on-escape
                    :opened="showItemTypeSheet" @actions:closed="showItemTypeSheet = false">
            <f7-actions-group>
                <f7-actions-label>{{ tt('Item Type') }}</f7-actions-label>
                <f7-actions-button v-for="def in itemDefinitions" :key="def.id"
                                   :bold="form.itemDefinitionId === def.id"
                                   @click="onItemTypeSelect(def.id)">
                    {{ def.name }}
                </f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>

        <!-- Enum Value Sheet -->
        <f7-actions close-by-outside-click close-on-escape
                    :opened="showEnumSheet" @actions:closed="showEnumSheet = false">
            <f7-actions-group>
                <f7-actions-label>{{ currentEnumField?.key || '' }}</f7-actions-label>
                <f7-actions-button v-for="opt in (currentEnumField?.options || [])" :key="opt"
                                   :bold="fieldValues[currentEnumField?.key || ''] === opt"
                                   @click="onEnumSelect(opt)">
                    {{ opt }}
                </f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>
    </f7-page>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import type { Router } from 'framework7/types';

import { useI18n } from '@/locales/helpers.ts';
import { useI18nUIComponents, showLoading, hideLoading } from '@/lib/ui/mobile.ts';

import api from '@/lib/services.ts';
import type { ItemDefinitionInfoResponse, ItemField } from '@/models/item_definition.ts';

const props = defineProps<{
    f7router: Router.Router;
    id?: string;
}>();

const { tt } = useI18n();
const { showToast } = useI18nUIComponents();

const saving = ref(false);
const isEditing = ref(false);
const editingId = ref('');
const showItemTypeSheet = ref(false);
const showEnumSheet = ref(false);
const currentEnumField = ref<ItemField | null>(null);

const itemDefinitions = ref<ItemDefinitionInfoResponse[]>([]);
const currentItemDefinition = ref<ItemDefinitionInfoResponse | null>(null);
const fieldValues = ref<Record<string, any>>({});

const form = ref({
    itemDefinitionId: '',
});

const selectedItemDefName = computed(() => {
    if (!form.value.itemDefinitionId) return '';
    return itemDefinitions.value.find(d => d.id === form.value.itemDefinitionId)?.name || '';
});

function isFieldValueSet(key: string): boolean {
    const v = fieldValues.value[key];
    return v !== null && v !== undefined && v !== '';
}

function onItemTypeSelect(itemDefId: string) {
    form.value.itemDefinitionId = itemDefId;
    showItemTypeSheet.value = false;
    loadItemDefinition(itemDefId);
}

function openEnumSheet(field: ItemField) {
    currentEnumField.value = field;
    showEnumSheet.value = true;
}

function onEnumSelect(value: string) {
    if (currentEnumField.value) {
        fieldValues.value[currentEnumField.value.key] = value;
    }
    showEnumSheet.value = false;
}

async function loadItemDefinition(itemDefId: string, existingValues?: Record<string, any>) {
    if (!itemDefId) {
        currentItemDefinition.value = null;
        fieldValues.value = {};
        return;
    }
    try {
        const resp = await api.getItemDefinition({ id: itemDefId });
        currentItemDefinition.value = resp.data.result;
        fieldValues.value = existingValues || {};
    } catch (error: any) {
        if (!error.processed) showToast(error.message || error);
    }
}

async function save() {
    if (!form.value.itemDefinitionId) {
        showToast(tt('Required'));
        return;
    }

    saving.value = true;
    showLoading();

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
                status: 'in_stock',
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
        hideLoading();
        props.f7router.back();
    } catch (error: any) {
        saving.value = false;
        hideLoading();
        if (!error.processed) showToast(error.message || error);
    }
}

function loadInventoryRecord(id: string) {
    showLoading();
    Promise.all([
        api.getInventoryRecord({ id }),
        api.getItemDefinitions(),
    ]).then(([recordResp, defsResp]) => {
        const record = recordResp.data.result;
        itemDefinitions.value = defsResp.data.result;
        editingId.value = record.id;
        isEditing.value = true;
        form.value.itemDefinitionId = record.itemDefinitionId;
        hideLoading();

        const def = itemDefinitions.value.find(d => d.id === record.itemDefinitionId);
        if (def) {
            currentItemDefinition.value = def;
            fieldValues.value = record.fieldValues?.values ? { ...record.fieldValues.values } : {};
        }
    }).catch(error => {
        hideLoading();
        if (!error.processed) showToast(error.message || error);
        props.f7router.back();
    });
}

onMounted(async () => {
    try {
        const defsResp = await api.getItemDefinitions();
        itemDefinitions.value = defsResp.data.result;
    } catch (error: any) {
        if (!error.processed) showToast(error.message || error);
    }

    if (props.id) {
        loadInventoryRecord(props.id);
    }
});
</script>
