const { computeSelectedTag, persistSelectedTag, restorePersistedTag } = require('../alarm-utils');

describe('Tag precedence and persistence', () => {
  beforeEach(() => {
    // jsdom provides a simple localStorage
    localStorage.clear();
  });

  test('URL param overrides persisted tag', () => {
    // Persist a tag 'outdoor'
    persistSelectedTag('alarm-selected-tag', 'outdoor');
    // Provide a URL search with ?tag=indoor
    const urlSearch = '?tag=indoor';
    const tagList = ['indoor','outdoor','garage'];

    const result = computeSelectedTag(urlSearch, restorePersistedTag('alarm-selected-tag'), tagList);
    expect(result.selectedTag).toBe('indoor');
    expect(result.newSearch).toBeNull();
  });

  test('Persisted tag applied when URL param absent and tag valid', () => {
    // Persist a tag 'garage'
    persistSelectedTag('alarm-selected-tag', 'garage');
    const urlSearch = '';
    const tagList = ['indoor','outdoor','garage'];

    const result = computeSelectedTag(urlSearch, restorePersistedTag('alarm-selected-tag'), tagList);
    expect(result.selectedTag).toBe('garage');
    // newSearch should include tag=garage
    expect(result.newSearch).toMatch(/tag=garage/);
  });

  test('Invalid persisted tag is ignored', () => {
    persistSelectedTag('alarm-selected-tag', 'invalid-tag');
    const urlSearch = '';
    const tagList = ['indoor','outdoor','garage'];

    const result = computeSelectedTag(urlSearch, restorePersistedTag('alarm-selected-tag'), tagList);
    expect(result.selectedTag).toBe('');
    expect(result.newSearch).toBeNull();
  });
});
