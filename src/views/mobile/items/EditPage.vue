<template>
    <f7-page>
        <f7-navbar>
            <f7-nav-left :back-link="tt('Back')" :class="{ 'disabled': saving }"></f7-nav-left>
            <f7-nav-title>{{ isEditing ? tt('Edit Item Definition') : tt('Add Item Definition') }}</f7-nav-title>
            <f7-nav-right :class="{ 'disabled': saving }">
                <f7-link icon-f7="checkmark_alt" @click="save" :class="{ 'disabled': saving }"></f7-link>
            </f7-nav-right>
        </f7-navbar>

        <f7-block-title class="margin-top">{{ tt('Basic Information') }}</f7-block-title>
        <f7-list strong inset>
            <f7-list-input type="text" :label="tt('Name')" :placeholder="tt('Name')"
                           :error-message="nameError" :error-message-force="!!nameError" clear-button
                           v-model:value="form.name" />

            <f7-list-item :title="tt('Item Icon')" link="#" no-chevron @click="showIconSheet = true">
                <template #media>
                    <ItemIcon icon-type="category" :icon-id="form.icon" />
                </template>
            </f7-list-item>

            <f7-list-input type="text" :label="tt('Pricing Expression')"
                           :placeholder="tt('Pricing Expression (e.g. weight * unit_price)')"
                           clear-button v-model:value="form.pricingExpr" />
            <f7-list-item v-if="validFieldKeys.length" class="field-key-chips">
                <template #title>
                    <span class="text-caption text-color-gray">{{ tt('Select field to insert into expression') }}:</span>
                </template>
                <template #after>
                    <f7-chip v-for="key in validFieldKeys" :key="key" :text="key"
                             class="margin-right-half" @click="insertFieldKey(key)" />
                </template>
            </f7-list-item>

            <f7-list-item :title="tt('Income Category')"
                          :after="incomeCategoryDisplayName"
                          link="#" @click="showIncomeCategorySheet = true" />
            <f7-list-item :title="tt('Expense Category')"
                          :after="expenseCategoryDisplayName"
                          link="#" @click="showExpenseCategorySheet = true" />
        </f7-list>

        <f7-block-title>
            {{ tt('Fields') }}
            <f7-link class="float-right" @click="addField">{{ tt('Add Field') }}</f7-link>
        </f7-block-title>
        <f7-list strong inset v-if="!form.fields.length">
            <f7-list-item :title="tt('No fields defined yet')" />
        </f7-list>

        <div v-for="(field, idx) in form.fields" :key="idx">
            <f7-list strong inset class="field-card-list">
                <f7-list-item :title="`${tt('Field')} ${idx + 1}`">
                    <template #after>
                        <f7-link color="red" @click="removeField(idx)">{{ tt('Delete') }}</f7-link>
                    </template>
                </f7-list-item>
                <f7-list-input type="text" :label="tt('Key')" :placeholder="tt('Key')"
                               clear-button v-model:value="field.key" />
                <f7-list-item :title="tt('Field Type')"
                              :after="fieldTypeLabels[field.fieldType] || ''"
                              link="#" @click="showFieldTypeSheetIdx = idx; showFieldTypeSheet = true" />

                <f7-list-input v-if="field.fieldType === 'number'" type="text" :label="tt('Unit')"
                               :placeholder="tt('Unit')" clear-button v-model:value="field.unit" />
                <f7-list-item v-if="field.fieldType === 'date'" :title="tt('Format')"
                              :after="dateFormatLabels[field.format || 'YYYY-MM-DD'] || 'YYYY-MM-DD'"
                              link="#" @click="showFormatSheetIdx = idx; showFormatSheet = true" />

                <f7-list-item :title="tt('Required')">
                    <template #after>
                        <f7-toggle :checked="field.required" @toggle:change="field.required = $event" />
                    </template>
                </f7-list-item>
                <f7-list-item :title="tt('Editable')">
                    <template #after>
                        <f7-toggle :checked="field.editable" @toggle:change="field.editable = $event" />
                    </template>
                </f7-list-item>
                <f7-list-item :title="tt('Participate in Naming')">
                    <template #after>
                        <f7-toggle :checked="field.participateInNaming" @toggle:change="field.participateInNaming = $event" />
                    </template>
                </f7-list-item>

                <div v-if="field.fieldType === 'enum'">
                    <f7-list-item :title="tt('Options')">
                        <template #after>
                            <f7-link @click="addOption(idx)">{{ tt('Add Option') }}</f7-link>
                        </template>
                    </f7-list-item>
                    <f7-list-input v-for="(_opt, optIdx) in (field.options || [])" :key="optIdx"
                                   type="text" :placeholder="`${tt('Option')} ${optIdx + 1}`"
                                   clear-button v-model:value="field.options![optIdx]">
                        <template #after>
                            <f7-link color="red" @click="removeOption(idx, optIdx)">
                                <f7-icon f7="xmark" size="16" />
                            </f7-link>
                        </template>
                    </f7-list-input>
                </div>
            </f7-list>
        </div>

        <div class="padding-horizontal padding-bottom">
            <f7-button fill @click="addField" :text="tt('Add Field')" />
        </div>

        <!-- Icon Selection Sheet -->
        <IconSelectionSheet icon-type="category"
                            :all-icon-infos="ALL_CATEGORY_ICONS"
                            :color="''"
                            v-model:show="showIconSheet"
                            v-model="form.icon" />

        <!-- Field Type Sheet -->
        <f7-actions close-by-outside-click close-on-escape
                    :opened="showFieldTypeSheet" @actions:closed="showFieldTypeSheet = false">
            <f7-actions-group>
                <f7-actions-label>{{ tt('Field Type') }}</f7-actions-label>
                <f7-actions-button v-for="opt in fieldTypeOptions" :key="opt.value"
                                   :bold="form.fields[showFieldTypeSheetIdx]?.fieldType === opt.value"
                                   @click="onFieldTypeSelect(opt.value)">
                    {{ fieldTypeLabels[opt.value] || opt.label }}
                </f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>

        <!-- Format Sheet -->
        <f7-actions close-by-outside-click close-on-escape
                    :opened="showFormatSheet" @actions:closed="showFormatSheet = false">
            <f7-actions-group>
                <f7-actions-label>{{ tt('Format') }}</f7-actions-label>
                <f7-actions-button v-for="opt in dateTimeFormatOptions" :key="opt.value"
                                   :bold="form.fields[showFormatSheetIdx]?.format === opt.value"
                                   @click="onFormatSelect(opt.value)">
                    {{ dateFormatLabels[opt.value] || opt.label }}
                </f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>

        <!-- Income Category Sheet -->
        <TwoColumnListItemSelectionSheet
            primary-key-field="id" primary-value-field="id" primary-title-field="name"
            primary-icon-field="icon" primary-icon-type="category" primary-color-field="color"
            primary-hidden-field="hidden" primary-sub-items-field="subCategories"
            secondary-key-field="id" secondary-value-field="id" secondary-title-field="name"
            secondary-icon-field="icon" secondary-icon-type="category" secondary-color-field="color"
            secondary-hidden-field="hidden"
            :enable-filter="true" :filter-placeholder="tt('Find category')"
            :filter-no-items-text="tt('No available category')"
            :items="incomeCategoryItems"
            v-model:show="showIncomeCategorySheet"
            v-model="form.incomeCategoryId" />

        <!-- Expense Category Sheet -->
        <TwoColumnListItemSelectionSheet
            primary-key-field="id" primary-value-field="id" primary-title-field="name"
            primary-icon-field="icon" primary-icon-type="category" primary-color-field="color"
            primary-hidden-field="hidden" primary-sub-items-field="subCategories"
            secondary-key-field="id" secondary-value-field="id" secondary-title-field="name"
            secondary-icon-field="icon" secondary-icon-type="category" secondary-color-field="color"
            secondary-hidden-field="hidden"
            :enable-filter="true" :filter-placeholder="tt('Find category')"
            :filter-no-items-text="tt('No available category')"
            :items="expenseCategoryItems"
            v-model:show="showExpenseCategorySheet"
            v-model="form.expenseCategoryId" />
    </f7-page>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import type { Router } from 'framework7/types';

import { useI18n } from '@/locales/helpers.ts';
import { useI18nUIComponents, showLoading, hideLoading } from '@/lib/ui/mobile.ts';

import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { CategoryType } from '@/core/category.ts';
import {
    getTransactionPrimaryCategoryName,
    getTransactionSecondaryCategoryName
} from '@/lib/category.ts';

import api from '@/lib/services.ts';
import type { TransactionCategory } from '@/models/transaction_category.ts';
import { ITEM_FIELD_TYPE_OPTIONS, ITEM_DATETIME_FORMAT_OPTIONS, ItemFieldType } from '@/models/item_definition.ts';

import ItemIcon from '@/components/mobile/ItemIcon.vue';
import IconSelectionSheet from '@/components/mobile/IconSelectionSheet.vue';
import TwoColumnListItemSelectionSheet from '@/components/mobile/TwoColumnListItemSelectionSheet.vue';
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

const props = defineProps<{
    f7route: Router.Route;
    f7router: Router.Router;
}>();

const id = props.f7route.query['id'] as string | undefined;

const { tt } = useI18n();
const { showToast } = useI18nUIComponents();

const transactionCategoriesStore = useTransactionCategoriesStore();
const incomeCategories = computed<TransactionCategory[]>(() =>
    transactionCategoriesStore.allTransactionCategories[CategoryType.Income] || []
);
const expenseCategories = computed<TransactionCategory[]>(() =>
    transactionCategoriesStore.allTransactionCategories[CategoryType.Expense] || []
);

const incomeCategoryItems = computed(() => incomeCategories.value as unknown as Record<string, unknown>[]);
const expenseCategoryItems = computed(() => expenseCategories.value as unknown as Record<string, unknown>[]);

const fieldTypeOptions = ITEM_FIELD_TYPE_OPTIONS;
const dateTimeFormatOptions = ITEM_DATETIME_FORMAT_OPTIONS;

const fieldTypeLabels = computed<Record<string, string>>(() => ({
    [ItemFieldType.NUMBER]: tt('Number'),
    [ItemFieldType.TEXT]: tt('Text'),
    [ItemFieldType.ENUM]: tt('Enum'),
    [ItemFieldType.DATE]: tt('Date'),
}));

const dateFormatLabels = computed<Record<string, string>>(() => ({
    'YYYY-MM-DD': tt('YYYY-MM-DD'),
    'YYYY-MM-DD HH:mm:ss': tt('YYYY-MM-DD HH:mm:ss'),
}));

const saving = ref(false);
const isEditing = ref(false);
const editingId = ref('');
const showIconSheet = ref(false);
const showFieldTypeSheet = ref(false);
const showFieldTypeSheetIdx = ref(0);
const showFormatSheet = ref(false);
const showFormatSheetIdx = ref(0);
const showIncomeCategorySheet = ref(false);
const showExpenseCategorySheet = ref(false);
const nameError = ref('');

const emptyField = (): MutableItemField => ({
    key: '', label: '', fieldType: 'number', required: true, editable: false,
    participateInNaming: false, unit: '', sortOrder: 0,
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

const incomeCategoryDisplayName = computed(() => {
    if (!form.value.incomeCategoryId) return '';
    const primary = getTransactionPrimaryCategoryName(form.value.incomeCategoryId, incomeCategories.value);
    const secondary = getTransactionSecondaryCategoryName(form.value.incomeCategoryId, incomeCategories.value);
    return secondary ? `${primary} > ${secondary}` : primary;
});

const expenseCategoryDisplayName = computed(() => {
    if (!form.value.expenseCategoryId) return '';
    const primary = getTransactionPrimaryCategoryName(form.value.expenseCategoryId, expenseCategories.value);
    const secondary = getTransactionSecondaryCategoryName(form.value.expenseCategoryId, expenseCategories.value);
    return secondary ? `${primary} > ${secondary}` : primary;
});

function onFieldTypeSelect(value: string) {
    const field = form.value.fields[showFieldTypeSheetIdx.value];
    if (field) {
        field.fieldType = value as MutableItemField['fieldType'];
        if (value === 'enum' && !field.options) field.options = [];
        if (value !== 'number') field.unit = '';
        if (value !== 'date') field.format = '';
    }
    showFieldTypeSheet.value = false;
}

function onFormatSelect(value: string) {
    const field = form.value.fields[showFormatSheetIdx.value];
    if (field) field.format = value;
    showFormatSheet.value = false;
}

function addField() {
    form.value.fields.push(emptyField());
}

function removeField(idx: number) {
    form.value.fields.splice(idx, 1);
}

function addOption(fieldIdx: number) {
    const field = form.value.fields[fieldIdx];
    if (field) {
        if (!field.options) field.options = [];
        field.options.push('');
    }
}

function removeOption(fieldIdx: number, optIdx: number) {
    const field = form.value.fields[fieldIdx];
    if (field?.options) field.options.splice(optIdx, 1);
}

function insertFieldKey(key: string) {
    form.value.pricingExpr += (form.value.pricingExpr ? ' ' : '') + key;
}

async function save() {
    if (!form.value.name.trim()) {
        nameError.value = tt('Required');
        return;
    }
    nameError.value = '';

    if (!form.value.incomeCategoryId) {
        showToast(tt('Income Category') + ' - ' + tt('Required'));
        return;
    }
    if (!form.value.expenseCategoryId) {
        showToast(tt('Expense Category') + ' - ' + tt('Required'));
        return;
    }

    saving.value = true;
    showLoading();

    try {
        const fieldSchema = { fields: form.value.fields.map((f, i) => ({ ...f, sortOrder: i })) };
        if (isEditing.value) {
            await api.modifyItemDefinition({
                id: editingId.value,
                name: form.value.name.trim(),
                icon: form.value.icon,
                fieldSchema,
                pricingExpr: form.value.pricingExpr,
                incomeCategoryId: form.value.incomeCategoryId || '0',
                expenseCategoryId: form.value.expenseCategoryId || '0',
            });
        } else {
            await api.addItemDefinition({
                name: form.value.name.trim(),
                icon: form.value.icon,
                fieldSchema,
                pricingExpr: form.value.pricingExpr,
                incomeCategoryId: form.value.incomeCategoryId || '0',
                expenseCategoryId: form.value.expenseCategoryId || '0',
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

function loadItemDefinition(id: string) {
    showLoading();
    api.getItemDefinition({ id }).then(resp => {
        const def = resp.data.result;
        editingId.value = def.id;
        isEditing.value = true;
        form.value = {
            name: def.name,
            icon: def.icon,
            pricingExpr: def.pricingExpr,
            incomeCategoryId: def.incomeCategoryId || '',
            expenseCategoryId: def.expenseCategoryId || '',
            fields: def.fieldSchema?.fields
                ? JSON.parse(JSON.stringify(def.fieldSchema.fields))
                : [],
        };
        hideLoading();
    }).catch(error => {
        hideLoading();
        if (!error.processed) showToast(error.message || error);
        props.f7router.back();
    });
}

onMounted(() => {
    transactionCategoriesStore.loadAllCategories({ force: false });
    if (id) {
        loadItemDefinition(id);
    }
});
</script>

<style scoped>
.field-card-list {
    margin-bottom: 8px;
}
.field-key-chips ::v-deep(.item-after) {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    max-width: 60%;
}
.margin-right-half {
    margin-right: 4px;
}
.float-right {
    float: right;
}
</style>
