const { getTriggeredBadgeClass } = require('../alarm-utils');

describe('getTriggeredBadgeClass', () => {
  test('returns ok for 0 or invalid', () => {
    expect(getTriggeredBadgeClass(0)).toBe('alarm-triggered-ok');
    expect(getTriggeredBadgeClass()).toBe('alarm-triggered-ok');
    expect(getTriggeredBadgeClass(null)).toBe('alarm-triggered-ok');
  });

  test('returns warn for 3', () => {
    expect(getTriggeredBadgeClass(3)).toBe('alarm-triggered-warn');
  });

  test('returns critical for 10', () => {
    expect(getTriggeredBadgeClass(10)).toBe('alarm-triggered-critical');
  });
});
