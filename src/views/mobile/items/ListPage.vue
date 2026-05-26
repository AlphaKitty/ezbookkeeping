<template>
    <f7-page :ptr="!loading" @ptr:refresh="reload" @page:afterin="onPageAfterIn">
        <f7-navbar>
            <f7-nav-left :back-link="tt('Back')" :class="{ 'disabled': loading }"></f7-nav-left>
            <f7-nav-title>{{ tt('Item Definitions') }}</f7-nav-title>
            <f7-nav-right :class="{ 'disabled': loading }">
                <f7-link icon-f7="plus" @click="add"></f7-link>
            </f7-nav-right>
        </f7-navbar>

        <f7-list strong inset dividers class="margin-top skeleton-text" v-if="loading">
            <f7-list-item :key="idx" v-for="idx in [1, 2, 3]">
                <template #media>
                    <f7-icon class="transaction-tag-icon" f7="cube"></f7-icon>
                </template>
                <template #title><div class="skeleton-text">Item Name</div></template>
                <template #footer><div class="skeleton-text">0 fields</div></template>
            </f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="margin-top" v-if="!loading && definitions.length === 0">
            <f7-list-item :title="tt('No data')"></f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="margin-top" v-if="!loading">
            <f7-list-item swipeout
                          :key="def.id"
                          v-for="def in definitions"
                          @click="edit(def)">
                <template #media>
                    <ItemIcon icon-type="category" :icon-id="def.icon" />
                </template>
                <template #title>{{ def.name }}</template>
                <template #footer>{{ tt('Fields') }}: {{ def.fieldSchema?.fields?.length || 0 }}</template>
                <template #after v-if="def.expensePricingExpr || def.incomePricingExpr">
                    <span class="text-caption text-color-gray text-no-wrap">
                        <span v-if="def.expensePricingExpr">↓{{ def.expensePricingExpr }}</span>
                        <span v-if="def.expensePricingExpr && def.incomePricingExpr"> </span>
                        <span v-if="def.incomePricingExpr">↑{{ def.incomePricingExpr }}</span>
                    </span>
                </template>
                <f7-swipeout-actions :right="true">
                    <f7-swipeout-button color="red" close @click="confirmDelete(def)">
                        <f7-icon f7="trash"></f7-icon>
                    </f7-swipeout-button>
                </f7-swipeout-actions>
            </f7-list-item>
        </f7-list>

        <f7-actions close-by-outside-click close-on-escape :opened="showDeleteSheet" @actions:closed="showDeleteSheet = false">
            <f7-actions-group>
                <f7-actions-label>{{ tt('Are you sure you want to delete this item?') }}</f7-actions-label>
                <f7-actions-button color="red" @click="doDelete">{{ tt('Delete') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>
    </f7-page>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import type { Router } from 'framework7/types';

import { useI18n } from '@/locales/helpers.ts';
import { useI18nUIComponents, showLoading, hideLoading } from '@/lib/ui/mobile.ts';

import api from '@/lib/services.ts';
import type { ItemDefinitionInfoResponse } from '@/models/item_definition.ts';

import ItemIcon from '@/components/mobile/ItemIcon.vue';

const props = defineProps<{
    f7router: Router.Router;
}>();

const { tt } = useI18n();
const { showToast, routeBackOnError } = useI18nUIComponents();

const loading = ref(false);
const loadingError = ref<unknown | null>(null);
const definitions = ref<ItemDefinitionInfoResponse[]>([]);
const showDeleteSheet = ref(false);
const deleteTarget = ref<ItemDefinitionInfoResponse | null>(null);
const hasNavigatedAway = ref(false);

function init() {
    loading.value = true;
    api.getItemDefinitions().then(resp => {
        definitions.value = resp.data.result;
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
    api.getItemDefinitions().then(resp => {
        definitions.value = resp.data.result;
        done?.();
        if (done) showToast('Item definitions have been updated');
    }).catch(error => {
        done?.();
        if (!error.processed) showToast(error.message || error);
    });
}

function add() {
    hasNavigatedAway.value = true;
    props.f7router.navigate('/item/definition/add');
}

function edit(def: ItemDefinitionInfoResponse) {
    hasNavigatedAway.value = true;
    props.f7router.navigate('/item/definition/edit?id=' + def.id);
}

function confirmDelete(def: ItemDefinitionInfoResponse) {
    deleteTarget.value = def;
    showDeleteSheet.value = true;
}

async function doDelete() {
    if (!deleteTarget.value) return;
    showDeleteSheet.value = false;
    showLoading();
    try {
        await api.deleteItemDefinition({ id: deleteTarget.value.id });
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
