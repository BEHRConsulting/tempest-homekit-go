/* eslint-env jest */
const path = require('path');
const { JSDOM } = require('jsdom');

describe('tag select UI interaction', () => {
  const scriptPath = path.resolve(__dirname, '../script.js');

  let _origWindow, _origDocument, _origLocalStorage, _origFetch;

  afterEach(() => {
    if (typeof _origWindow !== 'undefined') global.window = _origWindow; else delete global.window;
    if (typeof _origDocument !== 'undefined') global.document = _origDocument; else delete global.document;
    if (typeof _origLocalStorage !== 'undefined') global.localStorage = _origLocalStorage; else delete global.localStorage;
    if (typeof _origFetch !== 'undefined') global.fetch = _origFetch; else delete global.fetch;
    jest.resetModules();
  });

  test('changing the tag select updates localStorage and URL', () => {
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

    // Pre-populate persisted tag different from selection we'll choose
    window.localStorage.setItem('alarm-selected-tag', 'garage');

    const data = {
      enabled: true,
      totalAlarms: 2,
      enabledAlarms: 2,
      alarms: [ { enabled: true, name: 'A', tags: ['garage'] }, { enabled: true, name: 'B', tags: ['driveway'] } ]
    };

  expect(typeof updateAlarmStatus).toBe('function');
  updateAlarmStatus(data, { window, document: window.document });

    const sel = window.document.querySelector('select.alarm-tag-select');
    expect(sel).not.toBeNull();

    // Simulate user selecting 'driveway'
    sel.value = 'driveway';
    const event = new window.Event('change', { bubbles: true, cancelable: true });
    sel.dispatchEvent(event);

    // After change handler, localStorage should be updated and URL should include tag=driveway
    expect(window.localStorage.getItem('alarm-selected-tag')).toBe('driveway');
    expect(window.location.search).toMatch(/tag=driveway/);
  });
});
