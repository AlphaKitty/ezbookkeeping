<template>
    <f7-page :ptr="!loading" @ptr:refresh="reload" @page:afterin="onPageAfterIn">
        <f7-navbar>
            <f7-nav-left :back-link="tt('Back')" :class="{ 'disabled': loading }"></f7-nav-left>
            <f7-nav-title>{{ tt('Inventory Records') }}</f7-nav-title>
            <f7-nav-right :class="{ 'disabled': loading }">
                <f7-link icon-f7="plus" @click="add"></f7-link>
            </f7-nav-right>
        </f7-navbar>

        <!-- Item type filter chips -->
        <div class="padding-horizontal padding-top" v-if="itemDefinitions.length > 0">
            <f7-chip :text="tt('All Types')"
                     :class="{ 'chip-active': !activeItemDefId }"
                     class="margin-right-half margin-bottom-half"
                     @click="activeItemDefId = ''" />
            <f7-chip v-for="def in itemDefinitions" :key="def.id"
                     :text="def.name"
                     :class="{ 'chip-active': activeItemDefId === def.id }"
                     class="margin-right-half margin-bottom-half"
                     @click="activeItemDefId = activeItemDefId === def.id ? '' : def.id" />
        </div>

        <f7-list strong inset dividers class="margin-top skeleton-text" v-if="loading">
            <f7-list-item :key="idx" v-for="idx in [1, 2, 3]">
                <template #media>
                    <f7-icon f7="cube"></f7-icon>
                </template>
                <template #title><div class="skeleton-text">Record Name</div></template>
                <template #footer><div class="skeleton-text">Status · Item Type</div></template>
            </f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="margin-top" v-if="!loading && displayRecords.length === 0">
            <f7-list-item :title="tt('No data')"></f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="margin-top" v-if="!loading">
            <f7-list-item swipeout
                          :key="record.id"
                          v-for="record in displayRecords"
                          @click="edit(record)">
                <template #media>
                    <ItemIcon icon-type="category" :icon-id="getItemDefIcon(record.itemDefinitionId)" />
                </template>
                <template #title>{{ getRecordDisplayName(record) }}</template>
                <template #footer>{{ statusLabel(record.status) }} · {{ record.itemDefinitionName }}</template>
                <f7-swipeout-actions :right="true">
                    <f7-swipeout-button color="red" close @click="confirmDelete(record)">
                        <f7-icon f7="trash"></f7-icon>
                    </f7-swipeout-button>
                </f7-swipeout-actions>
            </f7-list-item>
        </f7-list>

        <f7-actions close-by-outside-click close-on-escape :opened="showDeleteSheet" @actions:closed="showDeleteSheet = false">
            <f7-actions-group>
                <f7-actions-label>{{ tt('Are you sure you want to delete this inventory record?') }}</f7-actions-label>
                <f7-actions-button color="red" @click="doDelete">{{ tt('Delete') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>
    </f7-page>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import type { Router } from 'framework7/types';

import { useI18n } from '@/locales/helpers.ts';
import { useI18nUIComponents, showLoading, hideLoading } from '@/lib/ui/mobile.ts';

import api from '@/lib/services.ts';
import type { InventoryRecordInfoResponse, InventoryStatus } from '@/models/inventory_record.ts';
import type { ItemDefinitionInfoResponse } from '@/models/item_definition.ts';

import ItemIcon from '@/components/mobile/ItemIcon.vue';

const props = defineProps<{
    f7router: Router.Router;
}>();

const { tt } = useI18n();
const { showToast, routeBackOnError } = useI18nUIComponents();

const loading = ref(false);
const loadingError = ref<unknown | null>(null);
const records = ref<InventoryRecordInfoResponse[]>([]);
const itemDefinitions = ref<ItemDefinitionInfoResponse[]>([]);
const activeItemDefId = ref<string>('');
const showDeleteSheet = ref(false);
const deleteTarget = ref<InventoryRecordInfoResponse | null>(null);
const hasNavigatedAway = ref(false);

const displayRecords = computed(() => {
    if (!activeItemDefId.value) return records.value;
    return records.value.filter(r => r.itemDefinitionId === activeItemDefId.value);
});

function getRecordDisplayName(record: InventoryRecordInfoResponse): string {
    const def = itemDefinitions.value.find(d => d.id === record.itemDefinitionId);
    const namingFields = def?.fieldSchema?.fields?.filter(f => f.participateInNaming) || [];
    const namingParts: string[] = [];
    if (record.fieldValues?.values) {
        for (const f of namingFields) {
            const val = record.fieldValues.values[f.key];
            if (val !== null && val !== undefined && val !== '') {
                const part = f.unit ? `${String(val)}${f.unit}` : String(val);
                namingParts.push(part);
            }
        }
    }
    if (record.itemDefinitionName) {
        return namingParts.length ? `${record.itemDefinitionName} - ${namingParts.join(' - ')}` : record.itemDefinitionName;
    }
    return `#${record.id}`;
}

function getItemDefIcon(itemDefinitionId: string): string {
    return itemDefinitions.value.find(d => d.id === itemDefinitionId)?.icon || '';
}

function statusLabel(status: InventoryStatus): string {
    return tt(status);
}

function init() {
    loading.value = true;
    Promise.all([
        api.getInventoryRecords(),
        api.getItemDefinitions(),
    ]).then(([recordsResp, defsResp]) => {
        records.value = recordsResp.data.result;
        itemDefinitions.value = defsResp.data.result;
        loading.value = false;
    }).catch(error => {
        loading.value = false;
        if (!error.processed) {
            loadingError.value = error;
            showToast(error.message || error);
        }
    });
}

function reload(done?: () => void) {
    Promise.all([
        api.getInventoryRecords(),
        api.getItemDefinitions(),
    ]).then(([recordsResp, defsResp]) => {
        records.value = recordsResp.data.result;
        itemDefinitions.value = defsResp.data.result;
        done?.();
        if (done) showToast('Inventory records have been updated');
    }).catch(error => {
        done?.();
        if (!error.processed) showToast(error.message || error);
    });
}

function add() {
    hasNavigatedAway.value = true;
    props.f7router.navigate('/inventory/record/add');
}

function edit(record: InventoryRecordInfoResponse) {
    hasNavigatedAway.value = true;
    props.f7router.navigate('/inventory/record/edit?id=' + record.id);
}

function confirmDelete(record: InventoryRecordInfoResponse) {
    deleteTarget.value = record;
    showDeleteSheet.value = true;
}

async function doDelete() {
    if (!deleteTarget.value) return;
    showDeleteSheet.value = false;
    showLoading();
    try {
        await api.deleteInventoryRecord({ id: deleteTarget.value.id });
        deleteTarget.value = null;
        hideLoading();
        reload();
    } catch (error: any) {
        hideLoading();
        if (!error.processed) showToast(error.message || error);
    }
}

function onPageAfterIn() {
    if (hasNavigatedAway.value && !loading.value) {
        reload();
    }
    hasNavigatedAway.value = false;
    routeBackOnError(props.f7router, loadingError);
}

init();
</script>

<style scoped>
.chip-active {
    --f7-chip-bg-color: var(--f7-theme-color);
    --f7-chip-text-color: #fff;
}
.margin-right-half {
    margin-right: 4px;
}
.margin-bottom-half {
    margin-bottom: 4px;
}
</style>
