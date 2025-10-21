/* eslint-env jest */
const { applyTagSelection } = require('../tag-handler.js');

describe('applyTagSelection', () => {
  test('sets tag and returns newSearch and storage set action', () => {
    const res = applyTagSelection('garage', '');
    expect(res).toBeDefined();
    expect(res.newSearch).toMatch(/tag=garage/);
    expect(res.storage).toEqual({ action: 'set', key: 'alarm-selected-tag', value: 'garage' });
  });

  test('clears tag and returns newSearch without tag and storage remove action', () => {
    const res = applyTagSelection('', '?tag=garage&foo=1');
    expect(res).toBeDefined();
    expect(res.newSearch).not.toMatch(/tag=/);
    expect(res.storage).toEqual({ action: 'remove', key: 'alarm-selected-tag' });
  });
});
