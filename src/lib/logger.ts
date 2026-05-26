import { isEnableDebug } from './settings.ts';

function logDebug(msg: string, obj?: unknown): void {
    if (isEnableDebug()) {
        if (obj) {
            console.debug('[企业记账 Debug] ' + msg, obj);
        } else {
            console.debug('[企业记账 Debug] ' + msg);
        }
    }
}

function logInfo(msg: string, obj?: unknown): void {
    if (obj) {
        console.info('[企业记账 Info] ' + msg, obj);
    } else {
        console.info('[企业记账 Info] ' + msg);
    }
}

function logWarn(msg: string, obj?: unknown): void {
    if (obj) {
        console.warn('[企业记账 Warn] ' + msg, obj);
    } else {
        console.warn('[企业记账 Warn] ' + msg);
    }
}

function logError(msg: string, obj?: unknown): void {
    if (obj) {
        console.error('[企业记账 Error] ' + msg, obj);
    } else {
        console.error('[企业记账 Error] ' + msg);
    }
}

export default {
    debug: logDebug,
    info: logInfo,
    warn: logWarn,
    error: logError
};
