// Helper for applying tag selection effects (URL + storage) in a testable way
function applyTagSelection(selectedValue, currentSearch = '', storageKey = 'alarm-selected-tag') {
  const params = new URLSearchParams(currentSearch || '');
  const result = { newSearch: null, storage: null };

  if (!selectedValue) {
    // clear tag
    params.delete('tag');
    result.newSearch = params.toString();
    result.storage = { action: 'remove', key: storageKey };
  } else {
    params.set('tag', selectedValue);
    result.newSearch = params.toString();
    result.storage = { action: 'set', key: storageKey, value: selectedValue };
  }

  return result;
}

if (typeof module !== 'undefined' && module.exports) {
  module.exports = { applyTagSelection };
} else {
  window.applyTagSelection = applyTagSelection;
}
