const { JSDOM } = require('jsdom');
const path = require('path');
const scriptPath = path.resolve(__dirname, '..', 'pkg', 'web', 'static', 'script.js');

const dom = new JSDOM(`<!doctype html><html><body>
  <div id="tempest-card"></div>
  <div id="alarm-list"><div class="alarm-list-header">Active Alarms</div></div>
</body></html>`, { url: 'http://localhost/?tag=indoor' });

global.window = dom.window;
global.document = dom.window.document;
global.localStorage = dom.window.localStorage;

delete require.cache[require.resolve(scriptPath)];
const script = require(scriptPath);
const updateAlarmStatus = script.updateAlarmStatus || dom.window.updateAlarmStatus;

const data = {
  enabled: true,
  totalAlarms: 2,
  enabledAlarms: 2,
  alarms: [
    { enabled: true, name: 'A', tags: ['outdoor'], condition: 'x', lastTriggered: '-', channels: ['email'], inCooldown: false, cooldownRemaining: 0, cooldown: 0, triggeredCount: 0 },
    { enabled: true, name: 'B', tags: ['indoor'], condition: 'y', lastTriggered: '-', channels: ['sms'], inCooldown: false, cooldownRemaining: 0, cooldown: 0, triggeredCount: 0 }
  ]
};

try {
  updateAlarmStatus(data);
  console.log('alarm-list innerHTML:\n', global.document.getElementById('alarm-list').innerHTML);
  const sel = global.document.querySelector('select.alarm-tag-select');
  console.log('select exists?', !!sel, sel && sel.outerHTML);
  const items = Array.from(global.document.querySelectorAll('.alarm-item')).map(el => el.outerHTML);
  console.log('alarm items:', items.length);
  items.forEach((it, i) => console.log(i, it));
} catch (e) {
  console.error('ERROR during updateAlarmStatus simulation:', e);
}
