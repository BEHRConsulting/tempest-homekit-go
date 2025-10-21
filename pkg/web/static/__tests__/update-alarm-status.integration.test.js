/* eslint-env jest */
const path = require('path');
const { JSDOM } = require('jsdom');

describe('updateAlarmStatus integration (real script)', () => {
  const scriptPath = path.resolve(__dirname, '../script.js');

  let _origWindow, _origDocument, _origLocalStorage, _origFetch;

  afterEach(() => {
    // Restore any globals we overwrote
    if (typeof _origWindow !== 'undefined') global.window = _origWindow; else delete global.window;
    if (typeof _origDocument !== 'undefined') global.document = _origDocument; else delete global.document;
    if (typeof _origLocalStorage !== 'undefined') global.localStorage = _origLocalStorage; else delete global.localStorage;
    if (typeof _origFetch !== 'undefined') global.fetch = _origFetch; else delete global.fetch;
    jest.resetModules();
  });

  test('URL tag param takes precedence over persisted localStorage', () => {
    const dom = new JSDOM('<!doctype html><html><body><div id="alarm-list"><div class="alarm-list-header">Active Alarms</div></div></body></html>', { url: 'http://localhost/?tag=indoor' });
    const { window } = dom;

    // Save originals and set JSDOM globals
    _origWindow = global.window;
    _origDocument = global.document;
    _origLocalStorage = global.localStorage;
    _origFetch = global.fetch;

    global.window = window;
    global.document = window.document;
    global.localStorage = window.localStorage;
    global.fetch = async () => ({ json: async () => ({}) });

    // Re-require the script to attach to this JSDOM window
    jest.resetModules();
    const script = require(scriptPath);
    const updateAlarmStatus = script.updateAlarmStatus || window.updateAlarmStatus;

    // Persisted different tag
    window.localStorage.setItem('alarm-selected-tag', 'outdoor');

    const data = {
      enabled: true,
      totalAlarms: 2,
      enabledAlarms: 2,
      alarms: [ { enabled: true, name: 'A', tags: ['outdoor'], condition: 'x', lastTriggered: '-', channels: ['email'], inCooldown: false, cooldownRemaining: 0, cooldown: 0, triggeredCount: 0 }, { enabled: true, name: 'B', tags: ['indoor'], condition: 'y', lastTriggered: '-', channels: ['sms'], inCooldown: false, cooldownRemaining: 0, cooldown: 0, triggeredCount: 0 } ]
    };

  expect(typeof updateAlarmStatus).toBe('function');
  updateAlarmStatus(data, { window, document: window.document });

    const sel = window.document.querySelector('select.alarm-tag-select');
    expect(sel).not.toBeNull();
    expect(sel.value).toBe('indoor');
  });

  test('Persisted tag applied and URL updated when no tag param', () => {
    const dom = new JSDOM('<!doctype html><html><body><div id="alarm-list"><div class="alarm-list-header">Active Alarms</div></div></body></html>', { url: 'http://localhost/' });
    const { window } = dom;

    _origWindow = global.window;
    _origDocument = global.document;
    _origLocalStorage = global.localStorage;
    _origFetch = global.fetch;

    global.window = window;
    global.document = window.document;
    global.localStorage = window.localStorage;
    global.fetch = async () => ({ json: async () => ({}) });

    jest.resetModules();
    const script = require(scriptPath);
    const updateAlarmStatus = script.updateAlarmStatus || window.updateAlarmStatus;

    window.localStorage.setItem('alarm-selected-tag', 'garage');

    const data = {
      enabled: true,
      totalAlarms: 1,
      enabledAlarms: 1,
      alarms: [ { enabled: true, name: 'C', tags: ['garage'], condition: 'z', lastTriggered: '-', channels: ['console'], inCooldown: false, cooldownRemaining: 0, cooldown: 0, triggeredCount: 0 } ]
    };

  expect(typeof updateAlarmStatus).toBe('function');
  updateAlarmStatus(data, { window, document: window.document });

    const sel = window.document.querySelector('select.alarm-tag-select');
    expect(sel).not.toBeNull();
    expect(sel.value).toBe('garage');
    expect(window.location.search).toMatch(/tag=garage/);
  });
});
