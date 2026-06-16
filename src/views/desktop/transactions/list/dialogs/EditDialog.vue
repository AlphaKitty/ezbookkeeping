<template>
    <v-dialog width="1000" :persistent="isTransactionModified" v-model="showState">
        <v-card class="pa-sm-1 pa-md-2">
            <template #title>
                <div class="d-flex align-center justify-center">
                    <div class="d-flex align-center">
                        <h4 class="text-h4">{{ tt(title) }}</h4>
                        <v-progress-circular indeterminate size="22" class="ms-2" v-if="loading"></v-progress-circular>
                    </div>
                    <v-spacer/>
                    <v-btn density="comfortable" color="default" variant="text" class="ms-2" :icon="true"
                           :disabled="loading || submitting" v-if="mode !== TransactionEditPageMode.View && (activeTab === 'basicInfo' || (activeTab === 'map' && isSupportGetGeoLocationByClick()))">
                        <v-icon :icon="mdiDotsVertical" />
                        <v-menu activator="parent">
                            <v-list v-if="activeTab === 'basicInfo'">
                                <v-list-item :prepend-icon="mdiSwapHorizontal"
                                             :title="tt('Swap Account')"
                                             v-if="transaction.type === TransactionType.Transfer"
                                             @click="swapTransactionData(true, false)"></v-list-item>
                                <v-list-item :prepend-icon="mdiSwapHorizontal"
                                             :title="tt('Swap Amount')"
                                             v-if="transaction.type === TransactionType.Transfer"
                                             @click="swapTransactionData(false, true)"></v-list-item>
                                <v-list-item :prepend-icon="mdiSwapHorizontal"
                                             :title="tt('Swap Account and Amount')"
                                             v-if="transaction.type === TransactionType.Transfer"
                                             @click="swapTransactionData(true, true)"></v-list-item>
                                <v-divider v-if="transaction.type === TransactionType.Transfer" />
                                <v-list-item :prepend-icon="mdiEyeOutline"
                                             :title="tt('Show Amount')"
                                             v-if="transaction.hideAmount" @click="transaction.hideAmount = false"></v-list-item>
                                <v-list-item :prepend-icon="mdiEyeOffOutline"
                                             :title="tt('Hide Amount')"
                                             v-if="!transaction.hideAmount" @click="transaction.hideAmount = true"></v-list-item>
                            </v-list>
                            <v-list v-if="activeTab === 'map'">
                                <v-list-item key="setGeoLocationByClickMap" value="setGeoLocationByClickMap"
                                             :prepend-icon="mdiMapMarkerOutline"
                                             :disabled="!transaction.geoLocation" v-if="isSupportGetGeoLocationByClick()">
                                    <v-list-item-title class="cursor-pointer" @click="setGeoLocationByClickMap = !setGeoLocationByClickMap; geoMenuState = false">
                                        <div class="d-flex align-center">
                                            <span>{{ tt('Click on Map to Set Geographic Location') }}</span>
                                            <v-spacer/>
                                            <v-icon :icon="mdiCheck" v-if="setGeoLocationByClickMap" />
                                        </div>
                                    </v-list-item-title>
                                </v-list-item>
                            </v-list>
                        </v-menu>
                    </v-btn>
                </div>
            </template>
            <v-card-text class="d-flex flex-column flex-md-row flex-grow-1 overflow-y-auto">
                <div class="mb-4">
                    <v-tabs class="v-tabs-pill" direction="vertical" :class="{ 'readonly': type === TransactionEditPageType.Transaction && mode !== TransactionEditPageMode.Add }"
                            :disabled="loading || submitting" v-model="transaction.type">
                        <v-tab :value="TransactionType.Expense" :disabled="type === TransactionEditPageType.Transaction && mode !== TransactionEditPageMode.Add && transaction.type !== TransactionType.Expense" v-if="transaction.type !== TransactionType.ModifyBalance">
                            <span>{{ tt('Expense') }}</span>
                        </v-tab>
                        <v-tab :value="TransactionType.Income" :disabled="type === TransactionEditPageType.Transaction && mode !== TransactionEditPageMode.Add && transaction.type !== TransactionType.Income" v-if="transaction.type !== TransactionType.ModifyBalance">
                            <span>{{ tt('Income') }}</span>
                        </v-tab>
                        <v-tab :value="TransactionType.Transfer" :disabled="type === TransactionEditPageType.Transaction && mode !== TransactionEditPageMode.Add && transaction.type !== TransactionType.Transfer" v-if="transaction.type !== TransactionType.ModifyBalance">
                            <span>{{ tt('Transfer') }}</span>
                        </v-tab>
                        <v-tab :value="TransactionType.ModifyBalance" v-if="type === TransactionEditPageType.Transaction && transaction.type === TransactionType.ModifyBalance">
                            <span>{{ tt('Modify Balance') }}</span>
                        </v-tab>
                    </v-tabs>
                    <v-divider class="my-2"/>
                    <v-tabs direction="vertical" :disabled="loading || submitting" v-model="activeTab">
                        <v-tab value="basicInfo">
                            <span>{{ tt('Basic Information') }}</span>
                        </v-tab>
                        <v-tab value="map" :disabled="!transaction.geoLocation" v-if="type === TransactionEditPageType.Transaction && !!getMapProvider()">
                            <span>{{ tt('Location on Map') }}</span>
                        </v-tab>
                        <v-tab value="pictures" :disabled="mode !== TransactionEditPageMode.Add && mode !== TransactionEditPageMode.Edit && (!transaction.pictures || !transaction.pictures.length)" v-if="type === TransactionEditPageType.Transaction && isTransactionPicturesEnabled()">
                            <span>{{ tt('Pictures') }}</span>
                        </v-tab>
                    </v-tabs>
                </div>

                <v-window class="d-flex flex-grow-1 disable-tab-transition w-100-window-container ms-md-5"
                          v-model="activeTab">
                    <v-window-item value="basicInfo">
                        <v-form class="mt-2">
                            <v-row>
                                <v-col cols="12" v-if="type === TransactionEditPageType.Template && transaction instanceof TransactionTemplate">
                                    <v-text-field
                                        type="text"
                                        persistent-placeholder
                                        :disabled="loading || submitting"
                                        :label="tt('Template Name')"
                                        :placeholder="tt('Template Name')"
                                        v-model="transaction.name"
                                    />
                                </v-col>
                                <v-col cols="12" :md="transaction.type === TransactionType.Transfer ? 6 : 12">
                                    <amount-input class="transaction-edit-amount font-weight-bold"
                                                  :color="sourceAmountColor"
                                                  :currency="sourceAccountCurrency"
                                                  :show-currency="true"
                                                  :readonly="mode === TransactionEditPageMode.View || (linkInventory && hasInventoryPricingExpr)"
                                                  :disabled="loading || submitting"
                                                  :persistent-placeholder="true"
                                                  :hide="transaction.hideAmount"
                                                  :label="sourceAmountTitle"
                                                  :placeholder="tt(sourceAmountName)"
                                                  :enable-formula="mode !== TransactionEditPageMode.View && !(linkInventory && hasInventoryPricingExpr)"
                                                  v-model="transaction.sourceAmount"/>
                                </v-col>
                                <v-col cols="12" :md="6" v-if="transaction.type === TransactionType.Transfer">
                                    <amount-input class="transaction-edit-amount font-weight-bold" color="primary"
                                                  :currency="destinationAccountCurrency"
                                                  :show-currency="true"
                                                  :readonly="mode === TransactionEditPageMode.View"
                                                  :disabled="loading || submitting"
                                                  :persistent-placeholder="true"
                                                  :hide="transaction.hideAmount"
                                                  :label="transferInAmountTitle"
                                                  :placeholder="tt('Transfer In Amount')"
                                                  :enable-formula="mode !== TransactionEditPageMode.View"
                                                  v-model="transaction.destinationAmount"/>
                                </v-col>
                                <v-col cols="12" v-if="transaction.type === TransactionType.Expense || transaction.type === TransactionType.Income">
                                    <v-checkbox v-model="linkInventory" :label="tt('Link Inventory')" :disabled="mode === TransactionEditPageMode.View" density="compact" hide-details class="mb-2" @update:model-value="onLinkInventoryToggle"/>
                                    <p v-if="linkInventory && hasInventoryPricingExpr" class="text-caption text-primary mt-1">{{ tt('inventoryAmountAutoCalculated') }}</p>
                                    <template v-if="linkInventory">
                                        <v-autocomplete v-model="selectedInventoryRecordIds" :label="tt('Inventory Record')" :items="inventoryRecordOptions" item-title="name" item-value="id" :custom-filter="filterInventoryRecord" density="compact" variant="outlined" class="mb-3" :disabled="mode === TransactionEditPageMode.View" clearable auto-select-first multiple @update:model-value="onInventoryRecordsChange"/>
                                        <div v-if="selectedInventoryRecordIds.length" class="d-flex flex-column ga-2 mb-3">
                                            <div v-for="recordId in selectedInventoryRecordIds" :key="recordId" class="d-flex align-center ga-2">
                                                <v-chip density="comfortable" variant="tonal" color="primary" size="small" closable @click:close="removeSelectedInventoryRecord(recordId)" @click="goToInventoryRecord(recordId)">
                                                    {{ getInventoryRecordName(recordId) }}
                                                </v-chip>
                                                <input type="number" min="0"
                                                    class="inventory-qty-input"
                                                    :class="{ 'inventory-qty-exceed': isQuantityExceedsStock(recordId) }"
                                                    :value="movedQuantities[recordId] ?? 0"
                                                    @input="onMovedQuantityChange(recordId, $event)"
                                                    :placeholder="tt('Quantity')"
                                                    :aria-label="tt('Quantity')"/>
                                                <span class="text-caption" :class="isQuantityExceedsStock(recordId) ? 'text-error' : 'text-disabled'">{{ getInventoryRecordStockInfo(recordId) }}</span>
                                                <v-tooltip v-if="isQuantityExceedsStock(recordId)" location="top">
                                                    <template v-slot:activator="{ props }">
                                                        <v-icon v-bind="props" :icon="mdiAlertCircleOutline" color="error" size="16"/>
                                                    </template>
                                                    <span>{{ tt('Quantity exceeds available stock') }} ({{ getRecordStock(recordId) }})</span>
                                                </v-tooltip>
                                                <v-tooltip v-else-if="getRecordMissingFields(recordId).length" location="top">
                                                    <template v-slot:activator="{ props }">
                                                        <v-icon v-bind="props" :icon="mdiAlertCircleOutline" color="warning" size="16"/>
                                                    </template>
                                                    <span>{{ tt('Missing fields') }}: {{ getRecordMissingFieldsSummary(recordId) }}</span>
                                                </v-tooltip>
                                            </div>
                                        </div>
                                        <p v-if="hasQuantityExceedRecords" class="text-caption text-error mt-1">
                                            <v-icon :icon="mdiAlertCircleOutline" size="14" class="me-1"/>
                                            {{ tt('Some inventory record quantities exceed available stock') }}
                                        </p>
                                        <p v-if="hasIncompleteFieldRecords" class="text-caption text-warning mt-1">
                                            <v-icon :icon="mdiAlertCircleOutline" size="14" class="me-1"/>
                                            {{ tt('Some inventory record fields are missing') }}
                                        </p>

                                        <!-- 计算公式面板 -->
                                        <div v-if="inventoryCalcBreakdown.length" class="inventory-calc-panel mt-2">
                                            <div class="text-caption font-weight-bold mb-2">{{ tt('Price Calculation') }}</div>
                                            <div v-for="item in inventoryCalcBreakdown" :key="item.recordId" class="calc-item mb-3">
                                                <div class="d-flex align-center ga-1 mb-1">
                                                    <v-chip density="compact" size="x-small" variant="tonal" color="primary">{{ item.itemDefName }}</v-chip>
                                                    <span class="text-caption text-disabled">× {{ item.movedQty }}</span>
                                                </div>
                                                <div class="calc-formula text-caption">
                                                    <!-- 表达式计算 -->
                                                    <template v-if="item.calcMethod === 'expression'">
                                                        <div class="d-flex flex-wrap align-center ga-1">
                                                            <code class="calc-expr font-weight-bold">{{ item.expr }}</code>
                                                            <span class="text-disabled">=</span>
                                                            <code class="calc-expr">{{ item.substituted }}</code>
                                                        </div>
                                                        <div v-if="item.fieldSchema.length" class="calc-subs mt-1">
                                                            <span class="text-disabled">{{ tt('where') }}:</span>
                                                            <span v-for="f in item.fieldSchema" :key="f.key" class="calc-var ms-2">
                                                                <code>{{ f.key }}</code>=<span :class="{ 'text-warning': isCalcFieldMissing(item.fieldValues || {}, f.key) }">{{ formatCalcFieldValue(item.fieldValues[f.key]) }}</span>
                                                            </span>
                                                        </div>
                                                    </template>
                                                    <!-- 无表达式：数量×单价 -->
                                                    <template v-else>
                                                        <code class="calc-expr">{{ item.movedQty }}</code>
                                                        <span class="text-disabled">×</span>
                                                        <code class="calc-expr">{{ item.unitPrice }}</code>
                                                        <span v-if="item.unitPrice === 0" class="text-warning text-caption ms-1">(unitPrice = 0)</span>
                                                    </template>
                                                </div>
                                                <div class="calc-result text-caption mt-1">
                                                    = <span class="font-weight-bold" :class="item.amount === undefined ? 'text-error' : (item.amount === 0 ? 'text-warning' : 'text-primary')">
                                                        {{ item.amount !== undefined ? formatAmount(item.amount) : '—' }}
                                                    </span>
                                                </div>
                                            </div>
                                            <v-divider class="my-1"/>
                                            <div class="calc-total text-body-2 font-weight-bold">
                                                {{ tt('Total Amount') }}: {{ formatAmount(transaction.sourceAmount) }}
                                            </div>
                                        </div>
                                    </template>
                                </v-col>
                                <v-col cols="12" md="12" v-if="transaction.type === TransactionType.Expense">
                                    <v-tooltip :disabled="hasVisibleExpenseCategories" :text="hasVisibleExpenseCategories ? '' : tt('No secondary expense categories are available')">
                                        <template v-slot:activator="{ props }">
                                            <div v-bind="props" class="d-block">
                                                <two-column-select primary-key-field="id" primary-value-field="id" primary-title-field="name"
                                                                   primary-icon-field="icon" primary-icon-type="category" primary-color-field="color"
                                                                   primary-hidden-field="hidden" primary-sub-items-field="subCategories"
                                                                   secondary-key-field="id" secondary-value-field="id" secondary-title-field="name"
                                                                   secondary-icon-field="icon" secondary-icon-type="category" secondary-color-field="color"
                                                                   secondary-hidden-field="hidden"
                                                                   :readonly="mode === TransactionEditPageMode.View || expenseCategoryLocked"
                                                                   :disabled="loading || submitting || !hasVisibleExpenseCategories"
                                                                   :enable-filter="true" :filter-placeholder="tt('Find category')" :filter-no-items-text="tt('No available category')"
                                                                   :show-selection-primary-text="true"
                                                                   :custom-selection-primary-text="getTransactionPrimaryCategoryName(transaction.expenseCategoryId, allCategories[CategoryType.Expense])"
                                                                   :custom-selection-secondary-text="getTransactionSecondaryCategoryName(transaction.expenseCategoryId, allCategories[CategoryType.Expense])"
                                                                   :label="tt('Category')" :placeholder="tt('Category')"
                                                                   :items="allCategories[CategoryType.Expense] || []"
                                                                   v-model="transaction.expenseCategoryId">
                                                </two-column-select>
                                            </div>
                                        </template>
                                    </v-tooltip>
                                </v-col>
                                <v-col cols="12" md="12" v-if="transaction.type === TransactionType.Income">
                                    <v-tooltip :disabled="hasVisibleIncomeCategories" :text="hasVisibleIncomeCategories ? '' : tt('No secondary income categories are available')">
                                        <template v-slot:activator="{ props }">
                                            <div v-bind="props" class="d-block">
                                                <two-column-select primary-key-field="id" primary-value-field="id" primary-title-field="name"
                                                                   primary-icon-field="icon" primary-icon-type="category" primary-color-field="color"
                                                                   primary-hidden-field="hidden" primary-sub-items-field="subCategories"
                                                                   secondary-key-field="id" secondary-value-field="id" secondary-title-field="name"
                                                                   secondary-icon-field="icon" secondary-icon-type="category" secondary-color-field="color"
                                                                   secondary-hidden-field="hidden"
                                                                   :readonly="mode === TransactionEditPageMode.View || incomeCategoryLocked"
                                                                   :disabled="loading || submitting || !hasVisibleIncomeCategories"
                                                                   :enable-filter="true" :filter-placeholder="tt('Find category')" :filter-no-items-text="tt('No available category')"
                                                                   :show-selection-primary-text="true"
                                                                   :custom-selection-primary-text="getTransactionPrimaryCategoryName(transaction.incomeCategoryId, allCategories[CategoryType.Income])"
                                                                   :custom-selection-secondary-text="getTransactionSecondaryCategoryName(transaction.incomeCategoryId, allCategories[CategoryType.Income])"
                                                                   :label="tt('Category')" :placeholder="tt('Category')"
                                                                   :items="allCategories[CategoryType.Income] || []"
                                                                   v-model="transaction.incomeCategoryId">
                                                </two-column-select>
                                            </div>
                                        </template>
                                    </v-tooltip>
                                </v-col>
                                <v-col cols="12" md="12" v-if="transaction.type === TransactionType.Transfer">
                                    <v-tooltip :disabled="hasVisibleTransferCategories" :text="hasVisibleTransferCategories ? '' : tt('No secondary transfer categories are available')">
                                        <template v-slot:activator="{ props }">
                                            <div v-bind="props" class="d-block">
                                                <two-column-select primary-key-field="id" primary-value-field="id" primary-title-field="name"
                                                                   primary-icon-field="icon" primary-icon-type="category" primary-color-field="color"
                                                                   primary-hidden-field="hidden" primary-sub-items-field="subCategories"
                                                                   secondary-key-field="id" secondary-value-field="id" secondary-title-field="name"
                                                                   secondary-icon-field="icon" secondary-icon-type="category" secondary-color-field="color"
                                                                   secondary-hidden-field="hidden"
                                                                   :readonly="mode === TransactionEditPageMode.View"
                                                                   :disabled="loading || submitting || !hasVisibleTransferCategories"
                                                                   :enable-filter="true" :filter-placeholder="tt('Find category')" :filter-no-items-text="tt('No available category')"
                                                                   :show-selection-primary-text="true"
                                                                   :custom-selection-primary-text="getTransactionPrimaryCategoryName(transaction.transferCategoryId, allCategories[CategoryType.Transfer])"
                                                                   :custom-selection-secondary-text="getTransactionSecondaryCategoryName(transaction.transferCategoryId, allCategories[CategoryType.Transfer])"
                                                                   :label="tt('Category')" :placeholder="tt('Category')"
                                                                   :items="allCategories[CategoryType.Transfer] || []"
                                                                   v-model="transaction.transferCategoryId">
                                                </two-column-select>
                                            </div>
                                        </template>
                                    </v-tooltip>
                                </v-col>
                                <v-col cols="12" :md="transaction.type === TransactionType.Transfer ? 6 : 12">
                                    <v-tooltip :disabled="!!allVisibleAccounts.length" :text="allVisibleAccounts.length ? '' : tt('No available account')">
                                        <template v-slot:activator="{ props }">
                                            <div v-bind="props" class="d-block">
                                                <two-column-select primary-key-field="id" primary-value-field="category"
                                                                   primary-title-field="name" primary-footer-field="displayBalance"
                                                                   primary-icon-field="icon" primary-icon-type="account"
                                                                   primary-sub-items-field="accounts"
                                                                   :primary-title-i18n="true"
                                                                   secondary-key-field="id" secondary-value-field="id"
                                                                   secondary-title-field="name" secondary-footer-field="displayBalance"
                                                                   secondary-icon-field="icon" secondary-icon-type="account" secondary-color-field="color"
                                                                   :readonly="mode === TransactionEditPageMode.View"
                                                                   :disabled="loading || submitting || !allVisibleAccounts.length || (mode === TransactionEditPageMode.Edit && transaction.type === TransactionType.ModifyBalance)"
                                                                   :enable-filter="true" :filter-placeholder="tt('Find account')" :filter-no-items-text="tt('No available account')"
                                                                   :custom-selection-primary-text="sourceAccountName"
                                                                   :label="tt(sourceAccountTitle)"
                                                                   :placeholder="tt(sourceAccountTitle)"
                                                                   :items="allVisibleCategorizedAccounts"
                                                                   v-model="transaction.sourceAccountId">
                                                </two-column-select>
                                            </div>
                                        </template>
                                    </v-tooltip>
                                </v-col>
                                <v-col cols="12" md="6" v-if="transaction.type === TransactionType.Transfer">
                                    <v-tooltip :disabled="!!allVisibleAccounts.length" :text="allVisibleAccounts.length ? '' : tt('No available account')">
                                        <template v-slot:activator="{ props }">
                                            <div v-bind="props" class="d-block">
                                                <two-column-select primary-key-field="id" primary-value-field="category"
                                                                   primary-title-field="name" primary-footer-field="displayBalance"
                                                                   primary-icon-field="icon" primary-icon-type="account"
                                                                   primary-sub-items-field="accounts"
                                                                   :primary-title-i18n="true"
                                                                   secondary-key-field="id" secondary-value-field="id"
                                                                   secondary-title-field="name" secondary-footer-field="displayBalance"
                                                                   secondary-icon-field="icon" secondary-icon-type="account" secondary-color-field="color"
                                                                   :readonly="mode === TransactionEditPageMode.View"
                                                                   :disabled="loading || submitting || !allVisibleAccounts.length"
                                                                   :enable-filter="true" :filter-placeholder="tt('Find account')" :filter-no-items-text="tt('No available account')"
                                                                   :custom-selection-primary-text="destinationAccountName"
                                                                   :label="tt('Destination Account')"
                                                                   :placeholder="tt('Destination Account')"
                                                                   :items="allVisibleCategorizedAccounts"
                                                                   v-model="transaction.destinationAccountId">
                                                </two-column-select>
                                            </div>
                                        </template>
                                    </v-tooltip>
                                </v-col>
                                <v-col cols="12" md="6" v-if="type === TransactionEditPageType.Transaction">
                                    <date-time-select
                                        :readonly="mode === TransactionEditPageMode.View"
                                        :disabled="loading || submitting || (mode === TransactionEditPageMode.Edit && transaction.type === TransactionType.ModifyBalance)"
                                        :label="tt('Transaction Time')"
                                        :timezone-utc-offset="transaction.utcOffset"
                                        :model-value="transaction.time"
                                        @update:model-value="updateTransactionTime"
                                        @error="onShowDateTimeError" />
                                </v-col>
                                <v-col cols="12" md="6" v-if="type === TransactionEditPageType.Template && transaction instanceof TransactionTemplate && transaction.templateType === TemplateType.Schedule.type">
                                    <schedule-frequency-select
                                        :readonly="mode === TransactionEditPageMode.View"
                                        :disabled="loading || submitting"
                                        :label="tt('Scheduled Transaction Frequency')"
                                        v-model:type="transaction.scheduledFrequencyType"
                                        v-model="transaction.scheduledFrequency" />
                                </v-col>
                                <v-col cols="12" md="6" v-if="type === TransactionEditPageType.Transaction || (type === TransactionEditPageType.Template && transaction instanceof TransactionTemplate && transaction.templateType === TemplateType.Schedule.type)">
                                    <v-autocomplete
                                        class="transaction-edit-timezone"
                                        item-title="displayNameWithUtcOffset"
                                        item-value="name"
                                        auto-select-first
                                        persistent-placeholder
                                        :readonly="mode === TransactionEditPageMode.View"
                                        :disabled="loading || submitting || (mode === TransactionEditPageMode.Edit && transaction.type === TransactionType.ModifyBalance)"
                                        :label="tt('Transaction Timezone')"
                                        :placeholder="!transaction.timeZone && transaction.timeZone !== '' ? `(${transactionDisplayTimezone}) ${transactionTimezoneTimeDifference}` : tt('Timezone')"
                                        :items="allTimezones"
                                        :no-data-text="tt('No results')"
                                        :model-value="transaction.timeZone"
                                        @update:model-value="updateTransactionTimezone"
                                    >
                                        <template #selection="{ item }">
                                            <span class="text-truncate" v-if="transaction.timeZone || transaction.timeZone === ''">
                                                {{ item.title }}
                                            </span>
                                        </template>
                                    </v-autocomplete>
                                </v-col>
                                <v-col cols="12" md="6" v-if="type === TransactionEditPageType.Template && transaction instanceof TransactionTemplate && transaction.templateType === TemplateType.Schedule.type">
                                    <date-select
                                        :readonly="mode === TransactionEditPageMode.View"
                                        :disabled="loading || submitting"
                                        :clearable="true"
                                        :label="tt('Start Date')"
                                        :no-data-text="tt('No limit')"
                                        v-model="transaction.scheduledStartDate" />
                                </v-col>
                                <v-col cols="12" md="6" v-if="type === TransactionEditPageType.Template && transaction instanceof TransactionTemplate && transaction.templateType === TemplateType.Schedule.type">
                                    <date-select
                                        :readonly="mode === TransactionEditPageMode.View"
                                        :disabled="loading || submitting"
                                        :clearable="true"
                                        :label="tt('End Date')"
                                        :no-data-text="tt('No limit')"
                                        v-model="transaction.scheduledEndDate" />
                                </v-col>
                                <v-col cols="12" md="12" v-if="type === TransactionEditPageType.Transaction">
                                    <v-select
                                        persistent-placeholder
                                        :readonly="mode === TransactionEditPageMode.View"
                                        :disabled="loading || submitting"
                                        :label="tt('Geographic Location')"
                                        v-model="transaction"
                                        v-model:menu="geoMenuState"
                                    >
                                        <template #selection>
                                            <span class="cursor-pointer" v-if="transaction.geoLocation">{{ `(${formatCoordinate(transaction.geoLocation, coordinateDisplayType)})` }}</span>
                                            <span class="cursor-pointer" v-else-if="!transaction.geoLocation">{{ geoLocationStatusInfo }}</span>
                                        </template>

                                        <template #no-data>
                                            <v-list class="py-0">
                                                <v-list-item v-if="mode !== TransactionEditPageMode.View" @click="updateGeoLocation(true)">{{ tt('Update Geographic Location') }}</v-list-item>
                                                <v-list-item v-if="mode !== TransactionEditPageMode.View" @click="clearGeoLocation">{{ tt('Clear Geographic Location') }}</v-list-item>
                                            </v-list>
                                        </template>
                                    </v-select>
                                </v-col>
                                <v-col cols="12" md="12">
                                    <transaction-tag-auto-complete
                                        :readonly="mode === TransactionEditPageMode.View"
                                        :disabled="loading || submitting"
                                        :show-label="true"
                                        :allow-add-new-tag="true"
                                        v-model="transaction.tagIds"
                                        @tag:saving="onSavingTag"
                                    />
                                </v-col>
                                <v-col cols="12" md="12">
                                    <v-textarea
                                        type="text"
                                        persistent-placeholder
                                        rows="3"
                                        :readonly="mode === TransactionEditPageMode.View"
                                        :disabled="loading || submitting"
                                        :label="tt('Description')"
                                        :placeholder="tt('Your transaction description (optional)')"
                                        v-model="transaction.comment"
                                    />
                                </v-col>
                            </v-row>
                        </v-form>
                    </v-window-item>
                    <v-window-item value="map">
                        <v-row>
                            <v-col cols="12" md="12">
                                <map-view ref="map" map-class="transaction-edit-map-view"
                                          :enable-zoom-control="true" :geo-location="transaction.geoLocation"
                                          @click="updateSpecifiedGeoLocation">
                                    <template #error-title="{ mapSupported, mapDependencyLoaded }">
                                        <span class="text-subtitle-1" v-if="!mapSupported"><b>{{ tt('Unsupported Map Provider') }}</b></span>
                                        <span class="text-subtitle-1" v-else-if="!mapDependencyLoaded"><b>{{ tt('Cannot Initialize Map') }}</b></span>
                                    </template>
                                    <template #error-content>
                                        <p class="text-body-1">
                                            {{ tt('Please refresh the page and try again. If the error persists, ensure that the server\'s map settings are correctly configured.') }}
                                        </p>
                                    </template>
                                </map-view>
                            </v-col>
                        </v-row>
                    </v-window-item>
                    <v-window-item value="pictures">
                        <v-row class="transaction-pictures align-content-start" :class="{ 'readonly': submitting || uploadingPicture || removingPictureId }">
                            <v-col :key="picIdx" cols="6" md="3" v-for="(pictureInfo, picIdx) in transaction.pictures">
                                <v-avatar rounded="lg" variant="tonal" size="160"
                                          class="cursor-pointer transaction-picture"
                                          color="rgba(0,0,0,0)" @click="viewOrRemovePicture(pictureInfo)">
                                    <v-img :src="getTransactionPictureUrl(pictureInfo)">
                                        <template #placeholder>
                                            <div class="d-flex align-center justify-center fill-height bg-light-primary">
                                                <v-progress-circular color="grey-500" indeterminate size="48"></v-progress-circular>
                                            </div>
                                        </template>
                                        <template #error>
                                            <div class="d-flex align-center justify-center fill-height bg-light-primary">
                                                <span class="text-body-1">{{ tt('Failed to load image, please check whether the config "domain" and "root_url" are set correctly.') }}</span>
                                            </div>
                                        </template>
                                    </v-img>
                                    <div class="picture-control-icon" :class="{ 'show-control-icon': pictureInfo.pictureId === removingPictureId }">
                                        <v-icon size="64" :icon="mdiTrashCanOutline" v-if="(mode === TransactionEditPageMode.Add || mode === TransactionEditPageMode.Edit) && pictureInfo.pictureId !== removingPictureId"/>
                                        <v-progress-circular color="grey-500" indeterminate size="48" v-if="(mode === TransactionEditPageMode.Add || mode === TransactionEditPageMode.Edit) && pictureInfo.pictureId === removingPictureId"></v-progress-circular>
                                        <v-icon size="64" :icon="mdiFullscreen" v-if="mode !== TransactionEditPageMode.Add && mode !== TransactionEditPageMode.Edit"/>
                                    </div>
                                </v-avatar>
                            </v-col>
                            <v-col cols="6" md="3" v-if="canAddTransactionPicture">
                                <v-avatar rounded="lg" variant="tonal" size="160"
                                          class="transaction-picture transaction-picture-add"
                                          :class="{ 'enabled': !submitting, 'cursor-pointer': !submitting }"
                                          color="rgba(0,0,0,0)" @click="showOpenPictureDialog">
                                    <v-tooltip activator="parent" v-if="!submitting">{{ tt('Add Picture') }}</v-tooltip>
                                    <v-icon class="transaction-picture-add-icon" size="56" :icon="mdiImagePlusOutline" v-if="!uploadingPicture"/>
                                    <v-progress-circular color="grey-500" indeterminate size="48" v-if="uploadingPicture"></v-progress-circular>
                                </v-avatar>
                            </v-col>
                        </v-row>
                    </v-window-item>
                </v-window>
            </v-card-text>
            <v-card-text>
                <div class="w-100 d-flex justify-center flex-wrap mt-sm-1 mt-md-2 gap-4">
                    <v-tooltip :disabled="!(inputIsEmpty || hasQuantityExceedRecords)" :text="saveDisabledTooltip">
                        <template v-slot:activator="{ props }">
                            <div v-bind="props" class="d-inline-block">
                                <v-btn-group density="comfortable" v-if="mode === TransactionEditPageMode.Add || mode === TransactionEditPageMode.Edit">
                                    <v-btn color="primary" :disabled="inputIsEmpty || hasQuantityExceedRecords || loading || submitting" @click="save(AfterSaveAction.GoBack)">
                                        {{ tt(saveButtonTitle) }}
                                        <v-progress-circular indeterminate size="22" class="ms-2" v-if="submitting"></v-progress-circular>
                                    </v-btn>
                                    <v-btn color="primary" density="compact"
                                           :disabled="inputIsEmpty || hasQuantityExceedRecords || loading || submitting" :icon="true"
                                           v-if="type === TransactionEditPageType.Transaction && mode === TransactionEditPageMode.Add">
                                        <v-icon :icon="mdiMenuDown" size="24" />
                                        <v-menu activator="parent">
                                            <v-list>
                                                <v-list-item :title="tt(TransactionQuickAddButtonActionType.SaveAndAddNewTransaction.name)"
                                                             @click="save(AfterSaveAction.StayWithNewTransaction)"></v-list-item>
                                                <v-list-item :title="tt(TransactionQuickAddButtonActionType.SaveAndKeepCurrentData.name)"
                                                             @click="save(AfterSaveAction.StayWithCurrentTransaction)"></v-list-item>
                                            </v-list>
                                        </v-menu>
                                    </v-btn>
                                </v-btn-group>
                            </div>
                        </template>
                    </v-tooltip>
                    <v-btn-group variant="tonal" density="comfortable"
                                 v-if="mode === TransactionEditPageMode.View && transaction.type !== TransactionType.ModifyBalance">
                        <v-btn :disabled="loading || submitting"
                               @click="duplicate(false, false)">{{ tt('Duplicate') }}</v-btn>
                        <v-btn density="compact" :disabled="loading || submitting" :icon="true">
                            <v-icon :icon="mdiMenuDown" size="24" />
                            <v-menu activator="parent">
                                <v-list>
                                    <v-list-item :title="tt('Duplicate (With Time)')"
                                                 @click="duplicate(true, false)"></v-list-item>
                                    <v-list-item :title="tt('Duplicate (With Geographic Location)')"
                                                 @click="duplicate(false, true)"
                                                 v-if="transaction.geoLocation"></v-list-item>
                                    <v-list-item :title="tt('Duplicate (With Time and Geographic Location)')"
                                                 @click="duplicate(true, true)"
                                                 v-if="transaction.geoLocation"></v-list-item>
                                </v-list>
                            </v-menu>
                        </v-btn>
                    </v-btn-group>
                    <v-btn color="warning" variant="tonal" :disabled="loading || submitting"
                           v-if="mode === TransactionEditPageMode.View && originalTransactionEditable"
                           @click="edit">{{ tt('Edit') }}</v-btn>
                    <v-btn color="error" variant="tonal" :disabled="loading || submitting"
                           v-if="mode === TransactionEditPageMode.View && originalTransactionEditable" @click="remove">
                        {{ tt('Delete') }}
                        <v-progress-circular indeterminate size="22" class="ms-2" v-if="submitting"></v-progress-circular>
                    </v-btn>
                    <v-btn color="secondary" variant="tonal" :disabled="loading || submitting"
                           @click="cancel">{{ tt(cancelButtonTitle) }}</v-btn>
                </div>
            </v-card-text>
        </v-card>
    </v-dialog>

    <!-- Inventory Record Edit Dialog -->
    <v-dialog v-model="showInventoryEditDialog" width="640" persistent>
        <v-card class="pa-sm-1 pa-md-2">
            <template #title>
                <h4 class="text-h4">{{ tt('Edit Inventory Record') }}</h4>
            </template>
            <v-card-text>
                <template v-if="inventoryEditCurrentItemDef?.fieldSchema?.fields?.length">
                    <div class="text-subtitle-2 mb-3">{{ inventoryEditCurrentItemDef.name }}</div>
                    <v-row>
                        <v-col v-for="field in inventoryEditCurrentItemDef.fieldSchema.fields" :key="field.key" cols="12" :md="field.fieldType === 'text' ? 12 : 6">
                            <v-text-field v-if="field.fieldType === 'number'"
                                v-model.number="inventoryEditFieldValues[field.key]"
                                :label="field.key"
                                :suffix="field.unit"
                                type="number"
                                density="compact" variant="outlined"
                                :rules="field.required ? [required] : []"/>
                            <v-text-field v-else-if="field.fieldType === 'text'"
                                v-model="inventoryEditFieldValues[field.key]"
                                :label="field.key"
                                density="compact" variant="outlined"
                                :rules="field.required ? [required] : []"/>
                            <v-select v-else-if="field.fieldType === 'enum'"
                                v-model="inventoryEditFieldValues[field.key]"
                                :label="field.key"
                                :items="(field as any).options || []"
                                density="compact" variant="outlined"
                                :rules="field.required ? [required] : []"/>
                            <v-text-field v-else-if="field.fieldType === 'date'"
                                v-model="inventoryEditFieldValues[field.key]"
                                :label="field.key"
                                :type="(field as any).format === 'YYYY-MM-DD HH:mm:ss' ? 'datetime-local' : 'date'"
                                density="compact" variant="outlined"
                                :rules="field.required ? [required] : []"/>
                        </v-col>
                    </v-row>
                </template>
                <p v-else class="text-caption text-disabled mt-4">{{ tt('This item type has no custom fields defined') }}</p>
                <v-alert v-if="inventoryEditFormError" type="error" variant="tonal" density="compact" class="mt-4" closable @click:close="inventoryEditFormError = ''">{{ inventoryEditFormError }}</v-alert>
            </v-card-text>
            <v-card-actions>
                <v-spacer/>
                <v-btn variant="text" @click="showInventoryEditDialog = false">{{ tt('Cancel') }}</v-btn>
                <v-btn color="primary" variant="tonal" :loading="inventoryEditSaving" @click="onInventoryEditSave">{{ tt('Save') }}</v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>

    <confirm-dialog ref="confirmDialog"/>
    <snack-bar ref="snackbar" />
    <input ref="pictureInput" type="file" style="display: none" :accept="SUPPORTED_IMAGE_EXTENSIONS" @change="uploadPicture($event)" />
</template>

<script setup lang="ts">
import MapView from '@/components/common/MapView.vue';
import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';

import { ref, computed, useTemplateRef, watch, nextTick } from 'vue';

import { useI18n } from '@/locales/helpers.ts';
import {
    TransactionEditPageMode,
    TransactionEditPageType,
    GeoLocationStatus,
    AfterSaveAction,
    useTransactionEditPageBase
} from '@/views/base/transactions/TransactionEditPageBase.ts';

import { useSettingsStore } from '@/stores/setting.ts';
import { useUserStore } from '@/stores/user.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionTagsStore } from '@/stores/transactionTag.ts';
import { useTransactionsStore } from '@/stores/transaction.ts';
import { useTransactionTemplatesStore } from '@/stores/transactionTemplate.ts';

import api from '@/lib/services.ts';
import type { ItemDefinitionInfoResponse } from '@/models/item_definition.ts';
import type { InventoryRecordInfoResponse } from '@/models/inventory_record.ts';
import { evaluateExpressionToAmount } from '@/lib/evaluator.ts';

import type { Coordinate } from '@/core/coordinate.ts';
import { CategoryType } from '@/core/category.ts';
import { TransactionType, TransactionEditScopeType, TransactionQuickAddButtonActionType } from '@/core/transaction.ts';
import { TemplateType, ScheduledTemplateFrequencyType } from '@/core/template.ts';
import { KnownErrorCode } from '@/consts/api.ts';
import { SUPPORTED_IMAGE_EXTENSIONS } from '@/consts/file.ts';

import { TransactionTemplate } from '@/models/transaction_template.ts';
import type { TransactionPictureInfoBasicResponse } from '@/models/transaction_picture_info.ts';
import { Transaction } from '@/models/transaction.ts';

import {
    getTimezoneOffsetMinutes,
    getCurrentUnixTime
} from '@/lib/datetime.ts';
import { formatCoordinate } from '@/lib/coordinate.ts';
import { generateRandomUUID } from '@/lib/misc.ts';
import {
    getTransactionPrimaryCategoryName,
    getTransactionSecondaryCategoryName
} from '@/lib/category.ts';
import { type SetTransactionOptions } from '@/lib/transaction.ts';
import {
    isTransactionPicturesEnabled,
    getMapProvider
} from '@/lib/server_settings.ts';
import {
    isSupportGetGeoLocationByClick
} from '@/lib/map/index.ts';
import logger from '@/lib/logger.ts';

import {
    mdiDotsVertical,
    mdiEyeOffOutline,
    mdiEyeOutline,
    mdiSwapHorizontal,
    mdiMapMarkerOutline,
    mdiCheck,
    mdiMenuDown,
    mdiImagePlusOutline,
    mdiTrashCanOutline,
    mdiFullscreen,
    mdiAlertCircleOutline
} from '@mdi/js';

export interface TransactionEditOptions extends SetTransactionOptions {
    id?: string;
    templateType?: number;
    template?: TransactionTemplate;
    currentTransaction?: Transaction;
    currentTemplate?: TransactionTemplate;
    noTransactionDraft?: boolean;
}

interface TransactionEditResponse {
    message: string;
}

type MapViewType = InstanceType<typeof MapView>;
type ConfirmDialogType = InstanceType<typeof ConfirmDialog>;
type SnackBarType = InstanceType<typeof SnackBar>;

const props = defineProps<{
    type: TransactionEditPageType;
    persistent?: boolean;
    show?: boolean;
}>();

const { tt } = useI18n();

const {
    mode,
    isSupportGeoLocation,
    editId,
    addByTemplateId,
    duplicateFromId,
    clientSessionId,
    loading,
    submitting,
    submitted,
    uploadingPicture,
    geoLocationStatus,
    setGeoLocationByClickMap,
    transaction,
    defaultCurrency,
    coordinateDisplayType,
    allTimezones,
    allVisibleAccounts,
    allVisibleCategorizedAccounts,
    allCategories,
    firstVisibleAccountId,
    hasVisibleExpenseCategories,
    hasVisibleIncomeCategories,
    hasVisibleTransferCategories,
    canAddTransactionPicture,
    title,
    saveButtonTitle,
    cancelButtonTitle,
    sourceAmountName,
    sourceAmountTitle,
    sourceAccountTitle,
    transferInAmountTitle,
    sourceAccountName,
    destinationAccountName,
    sourceAccountCurrency,
    destinationAccountCurrency,
    transactionDisplayTimezone,
    transactionTimezoneTimeDifference,
    geoLocationStatusInfo,
    inputEmptyProblemMessage,
    inputIsEmpty,
    createNewTransactionModel,
    setTransactionModel,
    updateTransactionModelByAfterSaveAction,
    updateTransactionTime,
    updateTransactionTimezone,
    swapTransactionData,
    getTransactionPictureUrl
} = useTransactionEditPageBase(props.type);

const settingsStore = useSettingsStore();
const userStore = useUserStore();
const accountsStore = useAccountsStore();
const transactionCategoriesStore = useTransactionCategoriesStore();
const transactionTagsStore = useTransactionTagsStore();
const transactionsStore = useTransactionsStore();
const transactionTemplatesStore = useTransactionTemplatesStore();

const map = useTemplateRef<MapViewType>('map');
const confirmDialog = useTemplateRef<ConfirmDialogType>('confirmDialog');
const snackbar = useTemplateRef<SnackBarType>('snackbar');
const pictureInput = useTemplateRef<HTMLInputElement>('pictureInput');

const showState = ref<boolean>(false);
const activeTab = ref<string>('basicInfo');
const originalTransactionEditable = ref<boolean>(false);
const noTransactionDraft = ref<boolean>(false);
const geoMenuState = ref<boolean>(false);
const removingPictureId = ref<string>('');

const initOptions = ref<TransactionEditOptions | undefined>(undefined);

const linkInventory = ref(false);
const inventoryItemDefs = ref<ItemDefinitionInfoResponse[]>([]);
const inventoryItemDefsLoaded = ref(false);
const inventoryRecords = ref<InventoryRecordInfoResponse[]>([]);
const selectedInventoryRecordIds = ref<string[]>([]);
const movedQuantities = ref<Record<string, number>>({});
const inventoryRecordsLoaded = ref(false);

const showInventoryEditDialog = ref(false);
const editingInventoryRecord = ref<InventoryRecordInfoResponse | null>(null);
const inventoryEditFieldValues = ref<Record<string, any>>({});
const inventoryEditCurrentItemDef = ref<ItemDefinitionInfoResponse | null>(null);
const inventoryEditSaving = ref(false);
const inventoryEditFormError = ref('');

const inventoryRecordOptions = computed(() => {
    const sorted = [...inventoryRecords.value].sort((a, b) => b.createdUnixTime - a.createdUnixTime);
    return sorted.map(r => {
        const def = inventoryItemDefs.value.find(d => d.id === r.itemDefinitionId);
        const namingFields = def?.fieldSchema?.fields?.filter(f => f.participateInNaming) || [];
        const namingParts: string[] = [];
        const allFieldTexts: string[] = [];
        if (r.fieldValues?.values) {
            for (const f of namingFields) {
                const val = r.fieldValues.values[f.key];
                if (val !== null && val !== undefined && val !== '') {
                    const part = f.unit ? `${String(val)}${f.unit}` : String(val);
                    namingParts.push(part);
                }
            }
            for (const val of Object.values(r.fieldValues.values)) {
                if (val !== null && val !== undefined && val !== '') {
                    allFieldTexts.push(String(val));
                }
            }
        }
        const displayName = r.itemDefinitionName
            ? (namingParts.length ? `${r.itemDefinitionName} - ${namingParts.join(' - ')}` : r.itemDefinitionName)
            : `#${r.id}`;
        const searchText = [r.itemDefinitionName || '', ...allFieldTexts].join(' ').toLowerCase();
        return { id: r.id, name: displayName, searchText };
    });
});

function filterInventoryRecord(item: any, queryText: string): boolean {
    if (!queryText) return true;
    const q = queryText.toLowerCase();
    return item.raw.searchText.includes(q);
}

function getSelectedRecordItemDef(recordId: string): ItemDefinitionInfoResponse | undefined {
    const record = inventoryRecords.value.find(r => r.id === recordId);
    if (!record) return undefined;
    return inventoryItemDefs.value.find(d => d.id === record.itemDefinitionId);
}

const expenseCategoryLocked = computed(() => {
    if (!linkInventory.value) return false;
    for (const recordId of selectedInventoryRecordIds.value) {
        const def = getSelectedRecordItemDef(recordId);
        if (def?.expenseCategoryId && def.expenseCategoryId !== '0') return true;
    }
    return false;
});
const incomeCategoryLocked = computed(() => {
    if (!linkInventory.value) return false;
    for (const recordId of selectedInventoryRecordIds.value) {
        const def = getSelectedRecordItemDef(recordId);
        if (def?.incomeCategoryId && def.incomeCategoryId !== '0') return true;
    }
    return false;
});

function getPricingExpr(def: ItemDefinitionInfoResponse): string {
    if (transaction.value.type === TransactionType.Expense) return def.expensePricingExpr;
    if (transaction.value.type === TransactionType.Income) return def.incomePricingExpr;
    return '';
}

const hasInventoryPricingExpr = computed(() => {
    return selectedInventoryRecordIds.value.length > 0;
});

const hasIncompleteFieldRecords = computed(() => {
    return selectedInventoryRecordIds.value.some(id => getRecordMissingFields(id).length > 0);
});

function getRecordMissingFieldsSummary(recordId: string): string {
    const missing = getRecordMissingFields(recordId);
    if (!missing.length) return '';
    return missing.join(', ');
}

function getRecordStock(recordId: string): number {
    const record = inventoryRecords.value.find(r => r.id === recordId);
    return record?.quantity || 0;
}

function isQuantityExceedsStock(recordId: string): boolean {
    const movedQty = movedQuantities.value[recordId] || 0;
    const stock = getRecordStock(recordId);
    return movedQty > 0 && stock > 0 && movedQty > stock;
}

const hasQuantityExceedRecords = computed(() => {
    return selectedInventoryRecordIds.value.some(id => isQuantityExceedsStock(id));
});

const saveDisabledTooltip = computed(() => {
    if (hasQuantityExceedRecords.value) return tt('Some inventory record quantities exceed available stock');
    if (inputIsEmpty.value && inputEmptyProblemMessage.value) return tt(inputEmptyProblemMessage.value);
    return '';
});

function formatCalcFieldValue(val: unknown): string {
    if (val === null || val === undefined || val === '') return '0 ⚠';
    if (typeof val === 'number' && isNaN(val)) return '0 ⚠';
    return String(val);
}

function isCalcFieldMissing(fieldValues: Record<string, unknown>, key: string): boolean {
    const val = fieldValues[key];
    return val === null || val === undefined || val === '' || (typeof val === 'number' && isNaN(val));
}

function formatAmount(amount: number | undefined | null): string {
    if (amount === undefined || amount === null) return '—';
    return (amount / 100).toFixed(2);
}

let resolveFunc: ((response?: TransactionEditResponse) => void) | null = null;
let rejectFunc: ((reason?: unknown) => void) | null = null;

const sourceAmountColor = computed<string | undefined>(() => {
    if (transaction.value.type === TransactionType.Expense) {
        return 'expense';
    } else if (transaction.value.type === TransactionType.Income) {
        return 'income';
    } else if (transaction.value.type === TransactionType.Transfer) {
        return 'primary';
    }

    return undefined;
});

const isTransactionModified = computed<boolean>(() => {
    if (mode.value === TransactionEditPageMode.Add) {
        return transactionsStore.isTransactionDraftModified(transaction.value, initOptions.value?.amount, initOptions.value?.categoryId, initOptions.value?.accountId, initOptions.value?.tagIds, firstVisibleAccountId.value);
    } else if (mode.value === TransactionEditPageMode.Edit) {
        return true;
    } else {
        return false;
    }
});

function open(options: TransactionEditOptions): Promise<TransactionEditResponse | undefined> {
    addByTemplateId.value = null;
    duplicateFromId.value = null;
    resetInventoryState();
    showState.value = true;
    activeTab.value = 'basicInfo';
    loading.value = true;
    submitting.value = false;
    submitted.value = false;
    geoLocationStatus.value = null;
    setGeoLocationByClickMap.value = false;
    originalTransactionEditable.value = false;
    noTransactionDraft.value = options.noTransactionDraft || false;

    initOptions.value = options;

    const newTransaction = createNewTransactionModel(options.type);
    setTransactionModel(newTransaction, options, true);

    const promises: Promise<unknown>[] = [
        accountsStore.loadAllAccounts({ force: false }),
        transactionCategoriesStore.loadAllCategories({ force: false }),
        transactionTagsStore.loadAllTags({ force: false })
    ];

    if (props.type === TransactionEditPageType.Transaction) {
        if (options && options.id) {
            if (options.currentTransaction) {
                setTransactionModel(options.currentTransaction, options, true);
            }

            mode.value = TransactionEditPageMode.View;
            editId.value = options.id;

            promises.push(transactionsStore.getTransaction({ transactionId: editId.value }));
        } else {
            mode.value = TransactionEditPageMode.Add;
            editId.value = null;

            if (options.template) {
                setTransactionModel(options.template, options, false);
                addByTemplateId.value = options.template.id;
            } else if (!options.noTransactionDraft && (settingsStore.appSettings.autoSaveTransactionDraft === 'enabled' || settingsStore.appSettings.autoSaveTransactionDraft === 'confirmation') && transactionsStore.transactionDraft) {
                setTransactionModel(Transaction.ofDraft(transactionsStore.transactionDraft), options, false);
            }

            if (settingsStore.appSettings.autoGetCurrentGeoLocation
                && !geoLocationStatus.value && !transaction.value.geoLocation) {
                updateGeoLocation(false);
            }
        }
    } else if (props.type === TransactionEditPageType.Template) {
        const template = TransactionTemplate.createNewTransactionTemplate(transaction.value);
        template.name = '';

        if (options && options.templateType) {
            template.templateType = options.templateType;
        }

        if (template.templateType === TemplateType.Schedule.type) {
            template.scheduledFrequencyType = ScheduledTemplateFrequencyType.Disabled.type;
            template.scheduledFrequency = '';
        }

        transaction.value = template;

        if (options && options.id) {
            if (options.currentTemplate) {
                setTransactionModel(options.currentTemplate, options, false);
                (transaction.value as TransactionTemplate).fillFrom(options.currentTemplate);
            }

            mode.value = TransactionEditPageMode.Edit;
            editId.value = options.id;
            transaction.value.id = options.id;

            promises.push(transactionTemplatesStore.getTemplate({ templateId: editId.value }));
        } else {
            mode.value = TransactionEditPageMode.Add;
            editId.value = null;
            transaction.value.id = '';
        }
    }

    if (options.type &&
        options.type >= TransactionType.Income &&
        options.type <= TransactionType.Transfer) {
        transaction.value.type = options.type;
    }

    if (mode.value === TransactionEditPageMode.Add) {
        clientSessionId.value = generateRandomUUID();
    }

    Promise.all(promises).then(function (responses) {
        if (editId.value && !responses[3]) {
            if (rejectFunc) {
                if (props.type === TransactionEditPageType.Transaction) {
                    rejectFunc('Unable to retrieve transaction');
                } else if (props.type === TransactionEditPageType.Template) {
                    rejectFunc('Unable to retrieve template');
                }
            }

            return;
        }

        if (props.type === TransactionEditPageType.Transaction && options && options.id && responses[3] && responses[3] instanceof Transaction) {
            const transaction: Transaction = responses[3];
            setTransactionModel(transaction, options, true);
            originalTransactionEditable.value = transaction.editable;
            if (transaction.inventoryRecordIds && transaction.inventoryRecordIds.length) {
                loadInventoryDataForExistingTransaction(transaction.inventoryRecordIds, transaction.inventoryRecordAmounts);
            } else if (transaction.inventoryRecordId) {
                loadInventoryDataForExistingTransaction([transaction.inventoryRecordId], transaction.inventoryRecordAmounts);
            }
        } else if (props.type === TransactionEditPageType.Template && options && options.id && responses[3] && responses[3] instanceof TransactionTemplate) {
            const template: TransactionTemplate = responses[3];
            setTransactionModel(template, options, false);

            if (!(transaction.value instanceof TransactionTemplate)) {
                transaction.value = TransactionTemplate.createNewTransactionTemplate(transaction.value);
            }

            (transaction.value as TransactionTemplate).fillFrom(template);
        } else {
            setTransactionModel(null, options, true);
        }

        loading.value = false;
    }).catch(error => {
        logger.error('failed to load essential data for editing transaction', error);

        loading.value = false;
        showState.value = false;

        if (!error.processed) {
            if (rejectFunc) {
                rejectFunc(error);
            }
        }
    });

    return new Promise((resolve, reject) => {
        resolveFunc = resolve;
        rejectFunc = reject;
    });
}

async function loadInventoryItemDefs() {
    if (inventoryItemDefsLoaded.value) return;
    const resp = await api.getItemDefinitions();
    inventoryItemDefs.value = resp.data.result;
    inventoryItemDefsLoaded.value = true;
}

async function loadInventoryRecords() {
    if (inventoryRecordsLoaded.value) return;
    try {
        const resp = await api.getInventoryRecords();
        inventoryRecords.value = resp.data.result;
        inventoryRecordsLoaded.value = true;
    } catch (error: any) {
        logger.warn('failed to load inventory records', error);
    }
}

function onLinkInventoryToggle(enabled: boolean | null) {
    if (enabled) {
        loadInventoryItemDefs();
        loadInventoryRecords();
    } else {
        selectedInventoryRecordIds.value = [];
    }
}

async function onInventoryRecordsChange() {
    if (!selectedInventoryRecordIds.value.length) {
        transaction.value.sourceAmount = 0;
        movedQuantities.value = {};
        return;
    }
    await loadInventoryItemDefs();
    for (const recordId of selectedInventoryRecordIds.value) {
        if (!(recordId in movedQuantities.value)) {
            movedQuantities.value[recordId] = 0;
        }
    }
    applyCategoryFromSelectedRecords();
    recalcInventoryAmount();
}

function applyCategoryFromSelectedRecords() {
    for (const recordId of selectedInventoryRecordIds.value) {
        const def = getSelectedRecordItemDef(recordId);
        if (!def) continue;
        if (def.incomeCategoryId && def.incomeCategoryId !== '0') {
            transaction.value.incomeCategoryId = def.incomeCategoryId;
        }
        if (def.expenseCategoryId && def.expenseCategoryId !== '0') {
            transaction.value.expenseCategoryId = def.expenseCategoryId;
        }
    }
}

async function loadInventoryDataForExistingTransaction(inventoryRecordIds: string[], existingAmounts?: number[]) {
    try {
        await loadInventoryItemDefs();
        await loadInventoryRecords();
        selectedInventoryRecordIds.value = inventoryRecordIds;
        linkInventory.value = true;
        for (let i = 0; i < inventoryRecordIds.length; i++) {
            const recordId = inventoryRecordIds[i];
            if (recordId) {
                movedQuantities.value[recordId] = existingAmounts?.[i] ?? 0;
            }
        }
        applyCategoryFromSelectedRecords();
        recalcInventoryAmount();
    } catch (error: any) {
        logger.warn('failed to load inventory data for existing transaction', error);
    }
}

function getRecordMissingFields(recordId: string): string[] {
    const def = getSelectedRecordItemDef(recordId);
    if (!def?.fieldSchema?.fields?.length) return [];
    const record = inventoryRecords.value.find(r => r.id === recordId);
    if (!record) return [];
    const fieldValues = record.fieldValues?.values || {};
    const missing: string[] = [];
    for (const field of def.fieldSchema.fields) {
        const val = fieldValues[field.key];
        if (val === null || val === undefined || val === '' || (typeof val === 'number' && isNaN(val))) {
            missing.push(field.key);
        }
    }
    return missing;
}

interface InventoryCalcBreakdownItem {
    recordId: string;
    itemDefName: string;
    movedQty: number;
    expr: string;
    exprSource: 'expense' | 'income' | 'none';
    unitPrice: number;
    fieldValues: Record<string, unknown>;
    fieldSchema: { key: string; unit?: string }[];
    substituted: string;
    amount: number | undefined;
    calcMethod: 'expression' | 'unitPrice';
    hasMissingFields: boolean;
}

const inventoryCalcBreakdown = ref<InventoryCalcBreakdownItem[]>([]);

function recalcInventoryAmount() {
    if (!selectedInventoryRecordIds.value.length) {
        inventoryCalcBreakdown.value = [];
        return;
    }

    let totalAmount = 0;
    const breakdown: InventoryCalcBreakdownItem[] = [];

    for (const recordId of selectedInventoryRecordIds.value) {
        const record = inventoryRecords.value.find(r => r.id === recordId);
        if (!record) continue;

        const def = getSelectedRecordItemDef(recordId);
        if (!def) continue;

        const movedQty = movedQuantities.value[recordId] || 0;
        const expr = getPricingExpr(def);
        const fieldValues = record.fieldValues?.values || {};
        const fieldSchema = (def.fieldSchema?.fields || []).map(f => ({ key: f.key, unit: f.unit }));

        if (expr) {
            const usesMovedQuantity = expr.includes('movedQuantity');
            let substituted = expr.replace(/movedQuantity/g, String(movedQty));
            let hasMissingFields = false;

            for (const field of def.fieldSchema?.fields || []) {
                const val = fieldValues[field.key];
                if (val === null || val === undefined || val === '' || (typeof val === 'number' && isNaN(val))) {
                    hasMissingFields = true;
                }
                const replacement = (val !== null && val !== undefined && val !== '' && !(typeof val === 'number' && isNaN(val)))
                    ? String(val)
                    : '0';
                substituted = substituted.replace(new RegExp(field.key, 'g'), replacement);
            }

            let unitAmount = evaluateExpressionToAmount(substituted);
            let finalAmount: number | undefined;
            let displaySubstituted: string;

            if (usesMovedQuantity) {
                // 表达式已包含 movedQuantity，直接使用结果
                finalAmount = unitAmount;
                displaySubstituted = substituted;
            } else {
                // 表达式不含 movedQuantity，计算单价后乘以数量
                displaySubstituted = `${movedQty} × (${substituted})`;
                if (unitAmount !== undefined) {
                    finalAmount = unitAmount * movedQty;
                }
            }

            breakdown.push({
                recordId,
                itemDefName: def.name,
                movedQty,
                expr,
                exprSource: transaction.value.type === TransactionType.Expense ? 'expense' : 'income',
                unitPrice: record.unitPrice,
                fieldValues,
                fieldSchema,
                substituted: displaySubstituted,
                amount: finalAmount,
                calcMethod: 'expression',
                hasMissingFields,
            });

            if (finalAmount !== undefined) {
                totalAmount += finalAmount;
            }
        } else {
            const amount = movedQty * record.unitPrice;

            breakdown.push({
                recordId,
                itemDefName: def.name,
                movedQty,
                expr: '',
                exprSource: 'none',
                unitPrice: record.unitPrice,
                fieldValues,
                fieldSchema,
                substituted: `${movedQty} × ${record.unitPrice}`,
                amount,
                calcMethod: 'unitPrice',
                hasMissingFields: false,
            });

            totalAmount += amount;
        }
    }

    inventoryCalcBreakdown.value = breakdown;

    if (totalAmount !== 0) {
        transaction.value.sourceAmount = totalAmount;
    } else if (selectedInventoryRecordIds.value.length > 0) {
        transaction.value.sourceAmount = 0;
    }
}

function resetInventoryState() {
    linkInventory.value = false;
    selectedInventoryRecordIds.value = [];
    movedQuantities.value = {};
    inventoryCalcBreakdown.value = [];
    inventoryItemDefsLoaded.value = false;
    inventoryRecordsLoaded.value = false;
    transaction.value.inventoryRecordId = undefined;
    transaction.value.inventoryRecordIds = undefined;
    transaction.value.inventoryRecordAmounts = undefined;
    transaction.value.inventoryAction = undefined;
}

function save(afterAction: AfterSaveAction): void {
    const problemMessage = inputEmptyProblemMessage.value;

    if (problemMessage) {
        snackbar.value?.showMessage(problemMessage);
        return;
    }

    if (linkInventory.value && hasQuantityExceedRecords.value) {
        snackbar.value?.showMessage(tt('Some inventory record quantities exceed available stock'));
        return;
    }

    if (props.type === TransactionEditPageType.Transaction && (mode.value === TransactionEditPageMode.Add || mode.value === TransactionEditPageMode.Edit)) {
        const doSubmit = function () {
            submitting.value = true;

            transactionsStore.saveTransaction({
                transaction: transaction.value as Transaction,
                defaultCurrency: defaultCurrency.value,
                isEdit: mode.value === TransactionEditPageMode.Edit,
                clientSessionId: clientSessionId.value
            }).then(() => {
                submitting.value = false;
                submitted.value = true;

                if (mode.value === TransactionEditPageMode.Add && !noTransactionDraft.value && !addByTemplateId.value && !duplicateFromId.value) {
                    transactionsStore.clearTransactionDraft();
                }

                if (mode.value === TransactionEditPageMode.Add && (afterAction === AfterSaveAction.StayWithNewTransaction || afterAction === AfterSaveAction.StayWithCurrentTransaction)) {
                    snackbar.value?.showMessage('You have added a new transaction');
                    updateTransactionModelByAfterSaveAction(afterAction, initOptions.value);
                    clientSessionId.value = generateRandomUUID();
                    resetInventoryState();
                } else {
                    if (resolveFunc) {
                        if (mode.value === TransactionEditPageMode.Add) {
                            resolveFunc({
                                message: 'You have added a new transaction'
                            });
                        } else if (mode.value === TransactionEditPageMode.Edit) {
                            resolveFunc({
                                message: 'You have saved this transaction'
                            });
                        }
                    }

                    showState.value = false;
                }
            }).catch(error => {
                submitting.value = false;

                if (error.error && (error.error.errorCode === KnownErrorCode.TransactionCannotCreateInThisTime || error.error.errorCode === KnownErrorCode.TransactionCannotModifyInThisTime)) {
                    confirmDialog.value?.open('You have set this time range to prevent editing transactions. Would you like to change the editable transaction range to All?').then(() => {
                        submitting.value = true;

                        userStore.updateUserTransactionEditScope({
                            transactionEditScope: TransactionEditScopeType.All.type
                        }).then(() => {
                            submitting.value = false;

                            snackbar.value?.showMessage('Your editable transaction range has been set to All');
                        }).catch(error => {
                            submitting.value = false;

                            if (!error.processed) {
                                snackbar.value?.showError(error);
                            }
                        });
                    });
                } else if (!error.processed) {
                    snackbar.value?.showError(error);
                }
            });
        };

        const saveWithOptionalInventory = async function () {
            if (linkInventory.value && selectedInventoryRecordIds.value.length) {
                transaction.value.inventoryRecordIds = selectedInventoryRecordIds.value;
                transaction.value.inventoryAction = transaction.value.type === TransactionType.Expense ? 'stock_in' : 'stock_out';
                const amounts: number[] = [];
                for (const recordId of selectedInventoryRecordIds.value) {
                    amounts.push(movedQuantities.value[recordId] || 0);
                }
                transaction.value.inventoryRecordAmounts = amounts;
            }

            if (transaction.value.sourceAmount === 0) {
                confirmDialog.value?.open('Are you sure you want to save this transaction with a zero amount?').then(() => {
                    doSubmit();
                });
            } else {
                doSubmit();
            }
        };

        saveWithOptionalInventory();
    } else if (props.type === TransactionEditPageType.Template && (mode.value === TransactionEditPageMode.Add || mode.value === TransactionEditPageMode.Edit)) {
        submitting.value = true;

        transactionTemplatesStore.saveTemplateContent({
            template: transaction.value as TransactionTemplate,
            isEdit: mode.value === TransactionEditPageMode.Edit,
            clientSessionId: clientSessionId.value
        }).then(() => {
            submitting.value = false;

            if (resolveFunc) {
                if (mode.value === TransactionEditPageMode.Add) {
                    resolveFunc({
                        message: 'You have added a new template'
                    });
                } else if (mode.value === TransactionEditPageMode.Edit) {
                    resolveFunc({
                        message: 'You have saved this template'
                    });
                }
            }

            showState.value = false;
        }).catch(error => {
            submitting.value = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    }
}

function duplicate(withTime?: boolean, withGeoLocation?: boolean): void {
    if (props.type !== TransactionEditPageType.Transaction || mode.value !== TransactionEditPageMode.View) {
        return;
    }

    editId.value = null;
    duplicateFromId.value = transaction.value.id;
    clientSessionId.value = generateRandomUUID();
    submitted.value = false;
    activeTab.value = 'basicInfo';
    transaction.value.id = '';
    transaction.value.inventoryRecordId = undefined;
    transaction.value.inventoryRecordIds = undefined;
    transaction.value.inventoryRecordAmounts = undefined;
    selectedInventoryRecordIds.value = [];

    if (!withTime) {
        transaction.value.time = getCurrentUnixTime();
        transaction.value.timeZone = settingsStore.appSettings.timeZone;
        transaction.value.utcOffset = getTimezoneOffsetMinutes(transaction.value.time, transaction.value.timeZone);
    }

    if (!withGeoLocation) {
        transaction.value.removeGeoLocation();
    }

    transaction.value.clearPictures();
    mode.value = TransactionEditPageMode.Add;
}

function edit(): void {
    if (props.type !== TransactionEditPageType.Transaction || mode.value !== TransactionEditPageMode.View) {
        return;
    }

    mode.value = TransactionEditPageMode.Edit;
}

function remove(): void {
    if (props.type !== TransactionEditPageType.Transaction || mode.value !== TransactionEditPageMode.View) {
        return;
    }

    confirmDialog.value?.open('Are you sure you want to delete this transaction?').then(() => {
        submitting.value = true;

        transactionsStore.deleteTransaction({
            transaction: transaction.value as Transaction,
            defaultCurrency: defaultCurrency.value
        }).then(() => {
            if (resolveFunc) {
                resolveFunc();
            }

            submitting.value = false;
            showState.value = false;
        }).catch(error => {
            submitting.value = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    });
}

function cancel(): void {
    const doClose = function () {
        if (props.type === TransactionEditPageType.Transaction && mode.value === TransactionEditPageMode.Add && submitted.value && resolveFunc) {
            resolveFunc({
                message: 'You have added a new transaction'
            });
        } else if (rejectFunc) {
            rejectFunc();
        }

        showState.value = false;
    };

    if (props.type !== TransactionEditPageType.Transaction || mode.value !== TransactionEditPageMode.Add || noTransactionDraft.value || addByTemplateId.value || duplicateFromId.value) {
        doClose();
        return;
    }

    if (settingsStore.appSettings.autoSaveTransactionDraft === 'confirmation') {
        if (transactionsStore.isTransactionDraftModified(transaction.value, initOptions.value?.amount, initOptions.value?.categoryId, initOptions.value?.accountId, initOptions.value?.tagIds, firstVisibleAccountId.value)) {
            confirmDialog.value?.open('Do you want to save this transaction draft?').then(() => {
                transactionsStore.saveTransactionDraft(transaction.value, initOptions.value?.amount, initOptions.value?.categoryId, initOptions.value?.accountId, initOptions.value?.tagIds, firstVisibleAccountId.value);
                doClose();
            }).catch(() => {
                transactionsStore.clearTransactionDraft();
                doClose();
            });
        } else {
            transactionsStore.clearTransactionDraft();
            doClose();
        }
    } else if (settingsStore.appSettings.autoSaveTransactionDraft === 'enabled') {
        transactionsStore.saveTransactionDraft(transaction.value, initOptions.value?.amount, initOptions.value?.categoryId, initOptions.value?.accountId, initOptions.value?.tagIds, firstVisibleAccountId.value);
        doClose();
    } else {
        doClose();
    }
}

function updateGeoLocation(forceUpdate: boolean): void {
    geoMenuState.value = false;

    if (!isSupportGeoLocation) {
        logger.warn('this browser does not support geo location');

        if (forceUpdate) {
            snackbar.value?.showMessage('Unable to retrieve current position');
        }
        return;
    }

    navigator.geolocation.getCurrentPosition(function (position) {
        if (!position || !position.coords) {
            logger.error('current position is null');
            geoLocationStatus.value = GeoLocationStatus.Error;

            if (forceUpdate) {
                snackbar.value?.showMessage('Unable to retrieve current position');
            }

            return;
        }

        geoLocationStatus.value = GeoLocationStatus.Success;

        transaction.value.setLatitudeAndLongitude(position.coords.latitude, position.coords.longitude);
    }, function (err) {
        logger.error('cannot retrieve current position', err);
        geoLocationStatus.value = GeoLocationStatus.Error;

        if (forceUpdate) {
            snackbar.value?.showMessage('Unable to retrieve current position');
        }
    });

    geoLocationStatus.value = GeoLocationStatus.Getting;
}

function updateSpecifiedGeoLocation(coordinate: Coordinate): void {
    if (isSupportGetGeoLocationByClick() && setGeoLocationByClickMap.value) {
        transaction.value.setLatitudeAndLongitude(coordinate.latitude, coordinate.longitude);
        map.value?.setMarkerPosition(transaction.value.geoLocation);
    }
}

function clearGeoLocation(): void {
    geoMenuState.value = false;
    geoLocationStatus.value = null;
    transaction.value.removeGeoLocation();
}

function showOpenPictureDialog(): void {
    if (!canAddTransactionPicture.value || submitting.value) {
        return;
    }

    pictureInput.value?.click();
}

function uploadPicture(event: Event): void {
    if (!event || !event.target) {
        return;
    }

    const el = event.target as HTMLInputElement;

    if (!el.files || !el.files.length || !el.files[0]) {
        return;
    }

    const pictureFile = el.files[0] as File;

    el.value = '';

    uploadingPicture.value = true;
    submitting.value = true;

    transactionsStore.uploadTransactionPicture({ pictureFile }).then(response => {
        transaction.value.addPicture(response);
        uploadingPicture.value = false;
        submitting.value = false;
    }).catch(error => {
        uploadingPicture.value = false;
        submitting.value = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function viewOrRemovePicture(pictureInfo: TransactionPictureInfoBasicResponse): void {
    if (mode.value !== TransactionEditPageMode.Add && mode.value !== TransactionEditPageMode.Edit) {
        window.open(getTransactionPictureUrl(pictureInfo), '_blank');
        return;
    }

    confirmDialog.value?.open('Are you sure you want to remove this transaction picture?').then(() => {
        removingPictureId.value = pictureInfo.pictureId;
        submitting.value = true;

        transactionsStore.removeUnusedTransactionPicture({ pictureInfo }).then(response => {
            if (response) {
                transaction.value.removePicture(pictureInfo);
            }

            removingPictureId.value = '';
            submitting.value = false;
        }).catch(error => {
            if (error.error && error.error.errorCode === KnownErrorCode.TransactionPictureNotFound) {
                transaction.value.removePicture(pictureInfo);
            } else if (!error.processed) {
                snackbar.value?.showError(error);
            }

            removingPictureId.value = '';
            submitting.value = false;
        });
    });
}

function onSavingTag(state: boolean): void {
    submitting.value = state;
}

function getInventoryRecordName(recordId: string): string {
    const opt = inventoryRecordOptions.value.find(o => o.id === recordId);
    return opt?.name || `#${recordId}`;
}

function getInventoryRecordStockInfo(recordId: string): string {
    const record = inventoryRecords.value.find(r => r.id === recordId);
    if (!record) return '';
    return (record.quantity || 0) + (record.unit ? ' ' + record.unit : '');
}

function onMovedQuantityChange(recordId: string, event: Event): void {
    const val = parseFloat((event.target as HTMLInputElement).value);
    movedQuantities.value[recordId] = isNaN(val) ? 0 : val;
    recalcInventoryAmount();
}

function removeSelectedInventoryRecord(recordId: string) {
    selectedInventoryRecordIds.value = selectedInventoryRecordIds.value.filter(id => id !== recordId);
    delete movedQuantities.value[recordId];
    recalcInventoryAmount();
}

function required(v: any): true | string {
    if (v === null || v === undefined || v === '') return tt('Required');
    if (typeof v === 'number' && isNaN(v)) return tt('Required');
    return true;
}

async function goToInventoryRecord(recordId: string) {
    const record = inventoryRecords.value.find(r => r.id === recordId);
    if (!record) return;
    editingInventoryRecord.value = record;
    inventoryEditFieldValues.value = record.fieldValues?.values ? { ...record.fieldValues.values } : {};
    inventoryEditFormError.value = '';

    let def = inventoryItemDefs.value.find(d => d.id === record.itemDefinitionId);
    if (!def) {
        try {
            const resp = await api.getItemDefinition({ id: record.itemDefinitionId });
            def = resp.data.result;
            inventoryItemDefs.value.push(def);
        } catch {
            inventoryEditFormError.value = tt('Failed to load item definition');
            return;
        }
    }
    inventoryEditCurrentItemDef.value = def;
    showInventoryEditDialog.value = true;
}

async function onInventoryEditSave() {
    const itemDef = inventoryEditCurrentItemDef.value;
    if (itemDef?.fieldSchema?.fields?.length) {
        for (const field of itemDef.fieldSchema.fields) {
            if (!field.required) continue;
            const v = inventoryEditFieldValues.value[field.key];
            if (v === null || v === undefined || v === '' || (typeof v === 'number' && isNaN(v))) {
                inventoryEditFormError.value = `${tt('Required')}: ${field.key}`;
                return;
            }
        }
    }

    const record = editingInventoryRecord.value;
    if (!record) return;

    inventoryEditSaving.value = true;
    try {
        const fieldValuesPayload = itemDef?.fieldSchema?.fields?.length
            ? { values: { ...inventoryEditFieldValues.value } }
            : null;

        await api.modifyInventoryRecord({
            id: record.id,
            itemDefinitionId: record.itemDefinitionId,
            warehouseId: record.warehouseId,
            fieldValues: fieldValuesPayload,
            quantity: record.quantity,
            unit: record.unit,
            unitPrice: record.unitPrice,
            transporter: record.transporter,
            batchNo: record.batchNo,
            status: record.status,
            comment: record.comment,
        });

        const updatedResp = await api.getInventoryRecord({ id: record.id });
        const updated = updatedResp.data.result;
        const idx = inventoryRecords.value.findIndex(r => r.id === record.id);
        if (idx >= 0) {
            inventoryRecords.value[idx] = updated;
        }

        showInventoryEditDialog.value = false;

        if (selectedInventoryRecordIds.value.includes(record.id)) {
            recalcInventoryAmount();
        }
    } catch (error: any) {
        if (!error.processed) {
            inventoryEditFormError.value = error.message || tt('Save failed');
        }
    } finally {
        inventoryEditSaving.value = false;
    }
}

function onShowDateTimeError(error: string): void {
    snackbar.value?.showError(error);
}

watch(activeTab, (newValue) => {
    if (newValue === 'map') {
        nextTick(() => {
            map.value?.initMapView();
        });
    }
});

watch(() => transaction.value.type, (newType) => {
    if (linkInventory.value && (newType === TransactionType.Expense || newType === TransactionType.Income)) {
        recalcInventoryAmount();
    }
});

watch(movedQuantities, () => {
    recalcInventoryAmount();
}, { deep: true });

defineExpose({
    open
});
</script>

<style>
.transaction-edit-amount .v-field__prepend-inner,
.transaction-edit-amount .v-field__append-inner,
.transaction-edit-amount .v-field__field > input {
    font-size: 1.25rem;
}

.transaction-edit-timezone.v-input input::placeholder {
    color: rgba(var(--v-theme-on-background), var(--v-high-emphasis-opacity)) !important;
    opacity: unset;
}

.transaction-edit-map-view {
    height: 220px;
}

@media (min-height: 630px) {
    .transaction-edit-map-view {
        height: 390px;
    }

    @media (min-width: 960px) {
        .transaction-pictures {
            min-height: 414px;
        }
    }
}

@media (min-height: 700px) {
    .transaction-edit-map-view {
        height: 460px;
    }

    @media (min-width: 960px) {
        .transaction-pictures {
            min-height: 484px;
        }
    }
}

@media (min-height: 780px) {
    .transaction-edit-map-view {
        height: 537px;
    }

    @media (min-width: 960px) {
        .transaction-pictures {
            min-height: 561px;
        }
    }
}

.transaction-picture .picture-control-icon {
    display: none;
    position: absolute;
    width: 100% !important;
    height: 100% !important;
    background-color: rgba(0, 0, 0, 0.4);
}

.transaction-picture .picture-control-icon > i.v-icon {
    background-color: transparent;
    color: rgba(255, 255, 255, 0.8);
}

.transaction-picture:hover .picture-control-icon,
.transaction-picture .picture-control-icon.show-control-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    vertical-align: middle;
}

.transaction-picture:hover .transaction-picture-placeholder {
    display: none;
}

.transaction-picture-add {
    border: 2px dashed rgba(var(--v-theme-grey-500));

    .transaction-picture-add-icon {
        color: rgba(var(--v-theme-grey-500));
    }
}

.transaction-picture-add.enabled:hover {
    border: 2px dashed rgba(var(--v-theme-grey-700));

    .transaction-picture-add-icon {
        color: rgba(var(--v-theme-grey-700));
    }
}

.inventory-qty-input {
    width: 80px;
    text-align: center;
    border: 1px solid rgba(var(--v-theme-on-surface), 0.22);
    border-radius: 4px;
    padding: 6px 8px;
    font-size: 14px;
    background: transparent;
    color: inherit;
    outline: none;
    transition: border-color 0.2s;
}

.inventory-qty-input:focus {
    border-color: rgb(var(--v-theme-primary));
    border-width: 2px;
    padding: 5px 7px;
}

.inventory-qty-input.inventory-qty-exceed {
    border-color: rgb(var(--v-theme-error));
    color: rgb(var(--v-theme-error));
}

.inventory-qty-input.inventory-qty-exceed:focus {
    border-color: rgb(var(--v-theme-error));
    border-width: 2px;
    padding: 5px 7px;
}

.inventory-qty-input::-webkit-inner-spin-button,
.inventory-qty-input::-webkit-outer-spin-button {
    opacity: 1;
}

.inventory-calc-panel {
    background: rgba(var(--v-theme-surface-variant), 0.3);
    border: 1px solid rgba(var(--v-theme-on-surface), 0.12);
    border-radius: 8px;
    padding: 10px 12px;
}

.inventory-calc-panel .calc-expr {
    background: rgba(var(--v-theme-primary), 0.08);
    color: rgb(var(--v-theme-primary));
    padding: 1px 6px;
    border-radius: 3px;
    font-size: 12px;
}

.inventory-calc-panel .calc-var code {
    background: rgba(var(--v-theme-on-surface), 0.06);
    padding: 0 4px;
    border-radius: 2px;
    font-size: 11px;
}

.inventory-calc-panel .calc-subs {
    font-size: 11px;
    line-height: 1.8;
}
</style>
