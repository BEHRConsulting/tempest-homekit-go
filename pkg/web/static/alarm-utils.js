// Small utility functions for alarm UI logic (pure functions - safe to test)
function getTriggeredBadgeClass(count) {
    const n = Number(count || 0);
    if (!Number.isFinite(n) || n <= 0) return 'alarm-triggered-ok';
    if (n > 5) return 'alarm-triggered-critical';
    return 'alarm-triggered-warn';
}

function persistSelectedTag(key, val) {
    try {
        if (!val) {
            localStorage.removeItem(key);
        } else {
            localStorage.setItem(key, val);
        }
    } catch (e) {
        // ignore storage errors
    }
}

function restorePersistedTag(key) {
    try {
        return localStorage.getItem(key);
    } catch (e) {
        return null;
    }
}

/**
 * Decide which tag should be selected based on URL and persisted value.
 * Returns an object { selectedTag, newSearch } where newSearch is null
 * when no URL update is required, or a string (without leading '?') when
 * the URL should be updated to include the selected tag.
 *
 * urlSearch: string (window.location.search)
 * persistedTag: string|null
 * tagList: array of available tag strings
 */
function computeSelectedTag(urlSearch, persistedTag, tagList) {
    const params = new URLSearchParams(urlSearch || '');
    const urlTag = params.get('tag');
    let selectedTag = urlTag || '';
    let newSearch = null;

    const tagSet = new Set(Array.isArray(tagList) ? tagList : []);

    if (urlTag) {
        // URL takes precedence, nothing to change
        selectedTag = urlTag;
        newSearch = null;
    } else if (!urlTag && persistedTag && tagSet.has(persistedTag)) {
        // No URL tag but persisted tag exists and is valid -> update URL
        selectedTag = persistedTag;
        params.set('tag', selectedTag);
        newSearch = params.toString();
    } else {
        selectedTag = '';
        newSearch = null;
    }

    return { selectedTag, newSearch };
}


// Export for test env (CommonJS & browser global)
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { getTriggeredBadgeClass, persistSelectedTag, restorePersistedTag, computeSelectedTag };
} else {
    window.getTriggeredBadgeClass = getTriggeredBadgeClass;
    window.persistSelectedTag = persistSelectedTag;
    window.restorePersistedTag = restorePersistedTag;
    window.computeSelectedTag = computeSelectedTag;
}
