let alarms = [];
let currentAlarm = null;
let allTags = [];
let selectedTags = [];

// ============================================
// Theme Switching System
// ============================================

// Load saved theme on page load
document.addEventListener('DOMContentLoaded', function() {
    const savedTheme = localStorage.getItem('selected-theme') || 'default';
    applyTheme(savedTheme);
    
    const themeSelect = document.getElementById('theme-select');
    if (themeSelect) {
        themeSelect.value = savedTheme;
        
        // Listen for theme changes
        themeSelect.addEventListener('change', function() {
            const newTheme = this.value;
            applyTheme(newTheme);
            localStorage.setItem('selected-theme', newTheme);
        });
    }
});

// Apply theme to document
function applyTheme(themeName) {
    const body = document.body;
    
    // Remove existing theme
    body.removeAttribute('data-theme');
    
    // Apply new theme (except for default)
    if (themeName !== 'default') {
        body.setAttribute('data-theme', themeName);
    }
}

async function init() {
    await loadAlarms();
    await loadTags();
    document.getElementById('searchName').addEventListener('input', filterAlarms);
    document.getElementById('filterTag').addEventListener('change', filterAlarms);
    document.getElementById('alarmForm').addEventListener('submit', handleSubmit);
    initTagSelector();
    
    // Update last update timestamp
    updateLastUpdateTimestamp();
}

async function loadAlarms() {
    const response = await fetch('/api/config?_=' + Date.now());
    const config = await response.json();
    alarms = config.alarms || [];
    renderAlarms();
}

async function loadTags() {
    const response = await fetch('/api/tags');
    allTags = await response.json();
    updateTagFilter();
}

function updateTagFilter() {
    const select = document.getElementById('filterTag');
    const currentValue = select.value;
    select.innerHTML = '<option value="">All Tags</option>';
    allTags.forEach(tag => {
        const option = document.createElement('option');
        option.value = tag;
        option.textContent = tag;
        if (tag === currentValue) option.selected = true;
        select.appendChild(option);
    });
}

function initTagSelector() {
    const searchInput = document.getElementById('tagSearchInput');
    const dropdown = document.getElementById('tagDropdown');
    
    searchInput.addEventListener('focus', () => {
        updateTagDropdown();
        dropdown.classList.add('active');
    });
    
    searchInput.addEventListener('input', (e) => {
        updateTagDropdown(e.target.value);
    });
    
    searchInput.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') {
            e.preventDefault();
            const value = searchInput.value.trim();
            if (value) {
                addTag(value);
                searchInput.value = '';
                updateTagDropdown();
            }
        }
    });
    
    // Close dropdown when clicking outside
    document.addEventListener('click', (e) => {
        if (!e.target.closest('.tag-dropdown-wrapper') && !e.target.closest('.tag-dropdown')) {
            dropdown.classList.remove('active');
        }
    });
}

function updateTagDropdown(searchTerm = '') {
    const dropdown = document.getElementById('tagDropdown');
    const searchLower = searchTerm.toLowerCase();
    
    // Filter available tags (not already selected)
    const availableTags = allTags.filter(tag => 
        !selectedTags.includes(tag) && 
        tag.toLowerCase().includes(searchLower)
    );
    
    dropdown.innerHTML = '';
    
    if (availableTags.length === 0 && !searchTerm) {
        dropdown.innerHTML = '<div class="tag-dropdown-empty">No tags available</div>';
        return;
    }
    
    // Show matching existing tags
    availableTags.forEach(tag => {
        const item = document.createElement('div');
        item.className = 'tag-dropdown-item';
        item.textContent = tag;
        item.addEventListener('click', () => {
            addTag(tag);
            document.getElementById('tagSearchInput').value = '';
            updateTagDropdown();
        });
        dropdown.appendChild(item);
    });
    
    // Show "add new tag" option if searching
    if (searchTerm && !allTags.includes(searchTerm) && !selectedTags.includes(searchTerm)) {
        const newTagItem = document.createElement('div');
        newTagItem.className = 'tag-dropdown-item new-tag';
        newTagItem.textContent = '+ Add new tag: "' + searchTerm + '"';
        newTagItem.addEventListener('click', () => {
            addTag(searchTerm);
            document.getElementById('tagSearchInput').value = '';
            updateTagDropdown();
        });
        dropdown.appendChild(newTagItem);
    }
    
    if (dropdown.children.length === 0 && searchTerm) {
        dropdown.innerHTML = '<div class="tag-dropdown-empty">No matching tags</div>';
    }
}

function addTag(tag) {
    const trimmedTag = tag.trim();
    if (!trimmedTag || selectedTags.includes(trimmedTag)) return;
    
    selectedTags.push(trimmedTag);
    
    // Add to allTags if it's new
    if (!allTags.includes(trimmedTag)) {
        allTags.push(trimmedTag);
        allTags.sort();
        updateTagFilter();
    }
    
    renderSelectedTags();
}

function removeTag(tag) {
    selectedTags = selectedTags.filter(t => t !== tag);
    renderSelectedTags();
    updateTagDropdown();
}

function renderSelectedTags() {
    const container = document.getElementById('selectedTags');
    
    if (selectedTags.length === 0) {
        container.innerHTML = '';
        return;
    }
    
    container.innerHTML = selectedTags.map(tag => 
        '<div class="selected-tag">' +
            '<span>' + tag + '</span>' +
            '<span class="remove-tag" onclick="removeTag(\'' + tag.replace(/'/g, "\\'") + '\')">×</span>' +
        '</div>'
    ).join('');
}

function filterAlarms() {
    renderAlarms();
}

function renderAlarms() {
    const nameFilter = document.getElementById('searchName').value.toLowerCase();
    const tagFilter = document.getElementById('filterTag').value;
    
    const filtered = alarms.filter(alarm => {
        if (nameFilter && !alarm.name.toLowerCase().includes(nameFilter)) return false;
        if (tagFilter && !alarm.tags.includes(tagFilter)) return false;
        return true;
    });
    
    const container = document.getElementById('alarmList');
    const emptyState = document.getElementById('emptyState');
    
    if (filtered.length === 0) {
        container.innerHTML = '';
        emptyState.style.display = 'block';
        return;
    }
    
    emptyState.style.display = 'none';
    container.innerHTML = filtered.map(alarm => {
        const enabledClass = alarm.enabled ? '' : 'disabled';
        const statusClass = alarm.enabled ? 'status-enabled' : 'status-disabled';
        const description = alarm.description ? '<div class="alarm-description">' + alarm.description + '</div>' : '';
        const tags = alarm.tags && alarm.tags.length ? '<div class="alarm-tags">' + alarm.tags.map(tag => '<span class="tag">' + tag + '</span>').join('') + '</div>' : '';
        const channels = alarm.channels ? alarm.channels.map(ch => ch.type).join(', ') : 'No channels';
        
        return '<div class="alarm-card ' + enabledClass + '">' +
            '<div class="alarm-header">' +
                '<div>' +
                    '<div class="alarm-name">' +
                        '<span class="status-indicator ' + statusClass + '"></span>' +
                        alarm.name +
                    '</div>' +
                    description +
                '</div>' +
            '</div>' +
            '<div class="alarm-condition">' + alarm.condition + '</div>' +
            tags +
            '<div class="alarm-channels">📢 ' + channels + '</div>' +
            '<div class="alarm-actions">' +
                '<button class="btn btn-primary" onclick="editAlarm(\'' + alarm.name + '\')">Edit</button>' +
                '<button class="btn btn-info btn-sm" onclick="showAlarmJSON(\'' + alarm.name + '\')">📄 JSON</button>' +
                '<button class="btn btn-danger" onclick="deleteAlarm(\'' + alarm.name + '\')">Delete</button>' +
            '</div>' +
        '</div>';
    }).join('');
}

function insertField(fieldName) {
    const textarea = document.getElementById('alarmCondition');
    const cursorPos = textarea.selectionStart;
    const textBefore = textarea.value.substring(0, cursorPos);
    const textAfter = textarea.value.substring(cursorPos);
    
    // Add space before if needed
    const needsSpaceBefore = textBefore.length > 0 && !textBefore.endsWith(' ') && !textBefore.endsWith('(');
    const prefix = needsSpaceBefore ? ' ' : '';
    
    textarea.value = textBefore + prefix + fieldName + textAfter;
    textarea.focus();
    textarea.setSelectionRange(cursorPos + prefix.length + fieldName.length, cursorPos + prefix.length + fieldName.length);
}

function insertVariable(textareaId, alternateId) {
    const select = event.target;
    const variable = select.value;
    if (!variable) return;
    
    // Determine which textarea to use
    let targetId = textareaId;
    if (alternateId) {
        // For email, determine if we should insert into subject or body based on focus
        const subject = document.getElementById(textareaId);
        const body = document.getElementById(alternateId);
        if (document.activeElement === body) {
            targetId = alternateId;
        }
    }
    
    const textarea = document.getElementById(targetId);
    const cursorPos = textarea.selectionStart;
    const textBefore = textarea.value.substring(0, cursorPos);
    const textAfter = textarea.value.substring(cursorPos);
    
    textarea.value = textBefore + variable + textAfter;
    textarea.focus();
    const newPos = cursorPos + variable.length;
    textarea.setSelectionRange(newPos, newPos);
    
    // Reset dropdown
    select.selectedIndex = 0;
}

function toggleMessageSections() {
    // Show/hide message sections based on selected delivery methods
    const consoleChecked = document.getElementById('deliveryConsole').checked;
    const syslogChecked = document.getElementById('deliverySyslog').checked;
    const oslogChecked = document.getElementById('deliveryOslog').checked;
    const eventlogChecked = document.getElementById('deliveryEventlog').checked;
    const emailChecked = document.getElementById('deliveryEmail').checked;
    const smsChecked = document.getElementById('deliverySMS').checked;
    const webhookChecked = document.getElementById('deliveryWebhook').checked;
    const csvChecked = document.getElementById('deliveryCSV').checked;
    const jsonChecked = document.getElementById('deliveryJSON').checked;
    
    // Message sections for each delivery method
    document.getElementById('consoleMessageSection').style.display = consoleChecked ? 'block' : 'none';
    document.getElementById('syslogMessageSection').style.display = syslogChecked ? 'block' : 'none';
    document.getElementById('oslogMessageSection').style.display = oslogChecked ? 'block' : 'none';
    document.getElementById('eventlogMessageSection').style.display = eventlogChecked ? 'block' : 'none';
    document.getElementById('emailMessageSection').style.display = emailChecked ? 'block' : 'none';
    document.getElementById('smsMessageSection').style.display = smsChecked ? 'block' : 'none';
    document.getElementById('webhookMessageSection').style.display = webhookChecked ? 'block' : 'none';
    document.getElementById('csvMessageSection').style.display = csvChecked ? 'block' : 'none';
    document.getElementById('jsonMessageSection').style.display = jsonChecked ? 'block' : 'none';
}

function showCreateModal() {
    currentAlarm = null;
    document.getElementById('alarmName').value = '';
    document.getElementById('alarmName').readOnly = false;
    document.getElementById('alarmDescription').value = '';
    document.getElementById('alarmCondition').value = '';
    document.getElementById('alarmCooldown').value = '1800';
    document.getElementById('alarmEnabled').checked = true;
    
    // Reset delivery methods to console only
    document.getElementById('deliveryConsole').checked = true;
    document.getElementById('deliverySyslog').checked = false;
    document.getElementById('deliveryOslog').checked = false;
    document.getElementById('deliveryEventlog').checked = false;
    document.getElementById('deliveryEmail').checked = false;
    document.getElementById('deliverySMS').checked = false;
    document.getElementById('deliveryWebhook').checked = false;
    document.getElementById('deliveryCSV').checked = false;
    document.getElementById('deliveryJSON').checked = false;
    
    // Set default messages with nice formatting
    // Console: Simple, clean terminal output
    document.getElementById('consoleMessage').value = `🚨 WEATHER ALARM TRIGGERED
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Alarm:       {{alarm_name}}
Description: {{alarm_description}}
Station:     {{station}}
Time:        {{timestamp}}
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
{{sensor_info}}`;
    
    // Syslog: Compact, structured format
    document.getElementById('syslogMessage').value = `tempest-alarm: {{alarm_name}} triggered - {{alarm_description}} | station={{station}} time={{timestamp}} | {{sensor_info}}`;
    
    // OSLog: Clean macOS logging format
    document.getElementById('oslogMessage').value = `[ALARM] {{alarm_name}}: {{alarm_description}} at {{station}} ({{timestamp}})`;
    
    // Event Log: Windows Event Viewer style
    document.getElementById('eventlogMessage').value = `Event Type: Warning
Source: Tempest HomeKit
Event ID: 1001
Description: Weather alarm condition detected

{{alarm_info}}

Station: {{station}}
Date/Time: {{timestamp}}

Sensor Data:
{{sensor_info}}

{{app_info}}`;
    
    // Email: Professional HTML-ready format
    // Will be populated from env defaults after modal opens
    document.getElementById('emailTo').value = '';
    document.getElementById('emailSubject').value = '⚠️ Weather Alert: {{alarm_name}}';
    document.getElementById('emailHtml').checked = true;
    document.getElementById('emailBody').value = `<h2 style="color: #d9534f;">⚠️ Weather Alarm Triggered</h2>

<div style="background: #f5f5f5; padding: 15px; border-left: 4px solid #d9534f; margin: 20px 0;">
    <h3 style="margin-top: 0;">{{alarm_name}}</h3>
    <p>{{alarm_description}}</p>
</div>

<table style="width: 100%; border-collapse: collapse; margin: 20px 0;">
    <tr style="background: #f9f9f9;">
        <td style="padding: 10px; border: 1px solid #ddd;"><strong>Station</strong></td>
        <td style="padding: 10px; border: 1px solid #ddd;">{{station}}</td>
    </tr>
    <tr>
        <td style="padding: 10px; border: 1px solid #ddd;"><strong>Time</strong></td>
        <td style="padding: 10px; border: 1px solid #ddd;">{{timestamp}}</td>
    </tr>
    <tr style="background: #f9f9f9;">
        <td style="padding: 10px; border: 1px solid #ddd;"><strong>Condition</strong></td>
        <td style="padding: 10px; border: 1px solid #ddd;">{{alarm_condition}}</td>
    </tr>
</table>

<h4>Current Sensor Readings:</h4>
<div style="background: #fff; padding: 15px; border: 1px solid #ddd;">
    {{sensor_info}}
</div>

<hr style="margin: 20px 0; border: none; border-top: 1px solid #ddd;">
<p style="color: #666; font-size: 12px;">{{app_info}}</p>`;
    
    // SMS: Very concise
    // Will be populated from env defaults after modal opens
    document.getElementById('smsTo').value = '';
    document.getElementById('smsMessage').value = `⚠️ {{alarm_name}} at {{station}} - {{timestamp}}. {{alarm_description}}`;
    
    // Webhook: JSON payload with alarm and sensor data
    document.getElementById('webhookUrl').value = '';
    document.getElementById('webhookMethod').value = 'POST';
    document.getElementById('webhookHeaders').value = `{
  "Content-Type": "application/json",
  "User-Agent": "Tempest-HomeKit-Alarm"
}`;
    document.getElementById('webhookBody').value = `{
  "alarm": {
    "name": "{{alarm_name}}",
    "description": "{{alarm_description}}",
    "condition": "{{alarm_condition}}",
    "tags": "{{alarm_tags}}"
  },
  "station": "{{station}}",
  "timestamp": "{{timestamp}}",
  "sensors": {
    "temperature_c": {{temperature}},
    "temperature_f": {{temperature_f}},
    "humidity": {{humidity}},
    "pressure_mb": {{pressure}},
    "wind_speed_ms": {{wind_speed}},
    "wind_gust_ms": {{wind_gust}},
    "wind_direction_deg": {{wind_direction}},
    "illuminance_lux": {{lux}},
    "uv_index": {{uv}},
    "rain_rate_mmh": {{rain_rate}},
    "rain_daily_mm": {{rain_daily}},
    "lightning_count": {{lightning_count}},
    "lightning_distance_km": {{lightning_distance}}
  },
  "app_info": "{{app_info}}"
}`;
    document.getElementById('webhookContentType').value = 'application/json';
    
    // CSV: Default path and message with timestamp, alarm info, and sensor data
    document.getElementById('csvPath').value = '/tmp/tempest-alarms.csv';
    document.getElementById('csvMaxDays').value = 30;
    document.getElementById('csvMessage').value = '{{alarm_name}},{{alarm_description}},{{temperature}},{{humidity}},{{pressure}},{{wind_speed}},{{lux}},{{uv}},{{rain_daily}},{{message}}';
    
    // JSON: Default path and message with timestamp, message, alarm info, and sensor info
    document.getElementById('jsonPath').value = '/tmp/tempest-alarms.json';
    document.getElementById('jsonMaxDays').value = 30;
    document.getElementById('jsonMessage').value = '{"timestamp": "{{timestamp}}", "message": "ALARM: {{alarm_name}} triggered", "alarm": {{alarm_info}}, "sensors": {{sensor_info}}}';
    
    selectedTags = [];
    renderSelectedTags();
    document.getElementById('tagSearchInput').value = '';
    
    toggleMessageSections();
    
    document.getElementById('editModal').classList.add('active');
    
    // Load environment defaults for email/SMS addresses
    loadEnvDefaults();
}

async function loadEnvDefaults() {
    try {
        const response = await fetch('/api/env-defaults');
        const defaults = await response.json();
        
        // Only set defaults if fields are empty (don't override user input)
        if (defaults.emailTo && !document.getElementById('emailTo').value) {
            document.getElementById('emailTo').value = defaults.emailTo;
        }
        if (defaults.smsTo && !document.getElementById('smsTo').value) {
            document.getElementById('smsTo').value = defaults.smsTo;
        }
    } catch (error) {
        console.warn('Failed to load environment defaults:', error);
    }
}

function editAlarm(name) {
    currentAlarm = alarms.find(a => a.name === name);
    if (!currentAlarm) return;
    
    document.getElementById('alarmName').value = currentAlarm.name;
    document.getElementById('alarmName').readOnly = false;
    document.getElementById('alarmDescription').value = currentAlarm.description || '';
    document.getElementById('alarmCondition').value = currentAlarm.condition;
    
    selectedTags = currentAlarm.tags || [];
    renderSelectedTags();
    document.getElementById('tagSearchInput').value = '';
    
    document.getElementById('alarmCooldown').value = currentAlarm.cooldown || 1800;
    document.getElementById('alarmEnabled').checked = currentAlarm.enabled;
    
    // Load delivery methods and messages from channels
    const channels = currentAlarm.channels || [];
    const channelTypes = channels.map(ch => ch.type);
    
    document.getElementById('deliveryConsole').checked = channelTypes.includes('console');
    document.getElementById('deliverySyslog').checked = channelTypes.includes('syslog');
    document.getElementById('deliveryOslog').checked = channelTypes.includes('oslog');
    document.getElementById('deliveryEventlog').checked = channelTypes.includes('eventlog');
    document.getElementById('deliveryEmail').checked = channelTypes.includes('email');
    document.getElementById('deliverySMS').checked = channelTypes.includes('sms');
    document.getElementById('deliveryWebhook').checked = channelTypes.includes('webhook');
    document.getElementById('deliveryCSV').checked = channelTypes.includes('csv');
    document.getElementById('deliveryJSON').checked = channelTypes.includes('json');
    
    // Load messages from channels
    channels.forEach(channel => {
        if (channel.type === 'console' && channel.template) {
            document.getElementById('consoleMessage').value = channel.template;
        } else if (channel.type === 'syslog' && channel.template) {
            document.getElementById('syslogMessage').value = channel.template;
        } else if (channel.type === 'oslog' && channel.template) {
            document.getElementById('oslogMessage').value = channel.template;
        } else if (channel.type === 'eventlog' && channel.template) {
            document.getElementById('eventlogMessage').value = channel.template;
        } else if (channel.type === 'email' && channel.email) {
            document.getElementById('emailTo').value = (channel.email.to || []).join(', ');
            document.getElementById('emailSubject').value = channel.email.subject || '';
            document.getElementById('emailBody').value = channel.email.body || '';
            document.getElementById('emailHtml').checked = channel.email.html || false;
        } else if (channel.type === 'sms' && channel.sms) {
            document.getElementById('smsTo').value = (channel.sms.to || []).join(', ');
            document.getElementById('smsMessage').value = channel.sms.message || '';
        } else if (channel.type === 'webhook' && channel.webhook) {
            document.getElementById('webhookUrl').value = channel.webhook.url || '';
            document.getElementById('webhookMethod').value = channel.webhook.method || 'POST';
            document.getElementById('webhookHeaders').value = channel.webhook.headers ? JSON.stringify(channel.webhook.headers, null, 2) : '';
            document.getElementById('webhookBody').value = channel.webhook.body || '';
            document.getElementById('webhookContentType').value = channel.webhook.content_type || 'application/json';
        } else if (channel.type === 'csv' && channel.csv) {
            document.getElementById('csvPath').value = channel.csv.path || '';
            document.getElementById('csvMaxDays').value = channel.csv.max_days || 30;
            document.getElementById('csvMessage').value = channel.csv.message || '';
        } else if (channel.type === 'json' && channel.json) {
            document.getElementById('jsonPath').value = channel.json.path || '';
            document.getElementById('jsonMaxDays').value = channel.json.max_days || 30;
            document.getElementById('jsonMessage').value = channel.json.message || '';
        }
    });
    
    toggleMessageSections();
    
    document.getElementById('editModal').classList.add('active');
}

function closeModal() {
    document.getElementById('editModal').classList.remove('active');
}

function closeJSONModal() {
    document.getElementById('jsonModal').classList.remove('active');
}

async function validateCondition() {
    const condition = document.getElementById('alarmCondition').value;
    const resultDiv = document.getElementById('validationResult');
    
    if (!condition || condition.trim() === '') {
        resultDiv.style.display = 'block';
        resultDiv.style.backgroundColor = '#fff3cd';
        resultDiv.style.color = '#856404';
        resultDiv.innerHTML = '⚠️ Please enter a condition to validate';
        return false;
    }
    
    try {
        const response = await fetch('/api/validate', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({ condition: condition })
        });
        
        const result = await response.json();
        resultDiv.style.display = 'block';
        
        if (result.valid) {
            resultDiv.style.backgroundColor = '#d4edda';
            resultDiv.style.color = '#155724';
            resultDiv.innerHTML = `✓ Valid condition!<br><strong>Meaning:</strong> ${result.paraphrase}`;
            return true;
        } else {
            resultDiv.style.backgroundColor = '#f8d7da';
            resultDiv.style.color = '#721c24';
            resultDiv.innerHTML = `✗ Invalid condition: ${result.error}`;
            return false;
        }
    } catch (error) {
        resultDiv.style.display = 'block';
        resultDiv.style.backgroundColor = '#f8d7da';
        resultDiv.style.color = '#721c24';
        resultDiv.innerHTML = `✗ Validation error: ${error.message}`;
        return false;
    }
}

async function validateJSONMessage() {
    const template = document.getElementById('jsonMessage').value;
    const resultDiv = document.getElementById('jsonValidationResult');
    
    if (!template || template.trim() === '') {
        resultDiv.style.display = 'block';
        resultDiv.style.backgroundColor = '#fff3cd';
        resultDiv.style.color = '#856404';
        resultDiv.innerHTML = '⚠️ Please enter a JSON template to validate';
        return false;
    }
    
    try {
        const response = await fetch('/api/validate-json', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({ template: template })
        });
        
        const result = await response.json();
        resultDiv.style.display = 'block';
        
        if (result.valid) {
            resultDiv.style.backgroundColor = '#d4edda';
            resultDiv.style.color = '#155724';
            resultDiv.innerHTML = `✓ Valid JSON template!<br><strong>Sample output:</strong><br><pre style="font-size: 11px; margin-top: 4px; word-wrap: break-word; white-space: pre-wrap;">${JSON.stringify(JSON.parse(result.expanded), null, 2)}</pre>`;
            return true;
        } else {
            resultDiv.style.backgroundColor = '#f8d7da';
            resultDiv.style.color = '#721c24';
            
            // Try to pretty-print even invalid JSON for better readability
            let formattedExpanded;
            try {
                formattedExpanded = JSON.stringify(JSON.parse(result.expanded), null, 2);
            } catch (parseError) {
                // If we can't parse it as JSON, show the raw text with some formatting
                formattedExpanded = result.expanded.replace(/\\n/g, '\n').replace(/\\t/g, '\t');
            }
            
            resultDiv.innerHTML = `✗ Invalid JSON template: ${result.error}<br><strong>Expanded result:</strong><br><pre style="font-size: 11px; margin-top: 4px; word-wrap: break-word; white-space: pre-wrap;">${formattedExpanded}</pre>`;
            return false;
        }
    } catch (error) {
        resultDiv.style.display = 'block';
        resultDiv.style.backgroundColor = '#f8d7da';
        resultDiv.style.color = '#721c24';
        resultDiv.innerHTML = `✗ Validation error: ${error.message}`;
        return false;
    }
}

async function showSampleJSON() {
    const template = document.getElementById('jsonMessage').value;
    
    if (!template || template.trim() === '') {
        showNotification('Please enter a JSON template first', 'error');
        return;
    }
    
    try {
        const response = await fetch('/api/validate-json', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({ template: template })
        });
        
        const result = await response.json();
        
        // Format the expanded result for display
        let formattedExpanded;
        let title;
        
        if (result.valid) {
            formattedExpanded = JSON.stringify(JSON.parse(result.expanded), null, 2);
            title = 'Sample JSON Output (Valid)';
        } else {
            title = 'Sample JSON Output (Invalid)';
            try {
                formattedExpanded = JSON.stringify(JSON.parse(result.expanded), null, 2);
            } catch (parseError) {
                formattedExpanded = result.expanded.replace(/\\n/g, '\n').replace(/\\t/g, '\t');
            }
        }
        
        displayJSON(formattedExpanded, title);
    } catch (error) {
        showNotification('Error generating sample JSON: ' + error.message, 'error');
    }
}

function showFullJSON() {
    const config = { alarms: alarms };
    displayJSON(config, 'Full Configuration JSON');
}

function showAlarmJSON(name) {
    const alarm = alarms.find(a => a.name === name);
    if (!alarm) return;
    displayJSON(alarm, 'Alarm: ' + name);
}

function displayJSON(data, title) {
    document.getElementById('jsonModalTitle').textContent = title;
    
    let jsonString;
    if (typeof data === 'string') {
        // If data is already a formatted JSON string, use it directly
        jsonString = data;
    } else {
        // If data is an object, stringify it
        jsonString = JSON.stringify(data, null, 2);
    }
    
    document.getElementById('jsonContent').textContent = jsonString;
    document.getElementById('jsonModal').classList.add('active');
}

async function copyJSON() {
    const jsonText = document.getElementById('jsonContent').textContent;
    try {
        await navigator.clipboard.writeText(jsonText);
        showNotification('JSON copied to clipboard!', 'success');
    } catch (err) {
        // Fallback for older browsers
        const textarea = document.createElement('textarea');
        textarea.value = jsonText;
        textarea.style.position = 'fixed';
        textarea.style.opacity = '0';
        document.body.appendChild(textarea);
        textarea.select();
        try {
            document.execCommand('copy');
            showNotification('JSON copied to clipboard!', 'success');
        } catch (e) {
            showNotification('Failed to copy JSON', 'error');
        }
        document.body.removeChild(textarea);
    }
}

async function handleSubmit(e) {
    e.preventDefault();
    
    // Validate condition before saving
    const isValid = await validateCondition();
    if (!isValid) {
        showNotification('Please fix the condition before saving', 'error');
        return;
    }
    
    // Validate JSON template if JSON delivery is selected
    if (document.getElementById('deliveryJSON').checked) {
        const jsonValid = await validateJSONMessage();
        if (!jsonValid) {
            showNotification('Please fix the JSON template before saving', 'error');
            return;
        }
    }
    
    // Build channels array from selected delivery methods
    const channels = [];
    
    if (document.getElementById('deliveryConsole').checked) {
        const template = document.getElementById('consoleMessage').value || '🚨 ALARM: {{alarm_name}}\nStation: {{station}}\nTime: {{timestamp}}';
        channels.push({ 
            type: 'console',
            template: template
        });
    }
    if (document.getElementById('deliverySyslog').checked) {
        const template = document.getElementById('syslogMessage').value || 'tempest-alarm: {{alarm_name}} - {{alarm_description}}';
        channels.push({ 
            type: 'syslog',
            template: template
        });
    }
    if (document.getElementById('deliveryOslog').checked) {
        const template = document.getElementById('oslogMessage').value || '[ALARM] {{alarm_name}}: {{alarm_description}}';
        channels.push({ 
            type: 'oslog',
            template: template
        });
    }
    if (document.getElementById('deliveryEventlog').checked) {
        const template = document.getElementById('eventlogMessage').value || 'Weather alarm: {{alarm_name}}';
        channels.push({ 
            type: 'eventlog',
            template: template
        });
    }
    if (document.getElementById('deliveryEmail').checked) {
        const emailTo = document.getElementById('emailTo').value;
        const emailSubject = document.getElementById('emailSubject').value || 'Tempest Alert: {{alarm_name}}';
        const emailBody = document.getElementById('emailBody').value || '{{alarm_info}}\n\n{{sensor_info}}';
        const emailHtml = document.getElementById('emailHtml').checked;
        
        channels.push({ 
            type: 'email',
            email: {
                to: emailTo ? emailTo.split(',').map(e => e.trim()).filter(e => e) : ['admin@example.com'],
                subject: emailSubject,
                body: emailBody,
                html: emailHtml
            }
        });
    }
    if (document.getElementById('deliverySMS').checked) {
        const smsTo = document.getElementById('smsTo').value;
        const smsMessage = document.getElementById('smsMessage').value;
        
        channels.push({ 
            type: 'sms',
            sms: {
                to: smsTo ? smsTo.split(',').map(p => p.trim()).filter(p => p) : ['+1234567890'],
                message: smsMessage || 'ALARM: {{alarm_name}} at {{timestamp}}'
            }
        });
    }
    if (document.getElementById('deliveryWebhook').checked) {
        const webhookUrl = document.getElementById('webhookUrl').value;
        const webhookMethod = document.getElementById('webhookMethod').value || 'POST';
        const webhookHeadersStr = document.getElementById('webhookHeaders').value;
        const webhookBody = document.getElementById('webhookBody').value;
        const webhookContentType = document.getElementById('webhookContentType').value || 'application/json';
        
        let webhookHeaders = {};
        if (webhookHeadersStr.trim()) {
            try {
                webhookHeaders = JSON.parse(webhookHeadersStr);
            } catch (e) {
                showNotification('Invalid JSON in webhook headers', 'error');
                return;
            }
        }
        
        channels.push({ 
            type: 'webhook',
            webhook: {
                url: webhookUrl,
                method: webhookMethod,
                headers: webhookHeaders,
                body: webhookBody,
                content_type: webhookContentType
            }
        });
    }
    
    if (document.getElementById('deliveryCSV').checked) {
        const csvPath = document.getElementById('csvPath').value;
        const csvMaxDays = parseInt(document.getElementById('csvMaxDays').value) || 30;
        const csvMessage = document.getElementById('csvMessage').value || '{{alarm_name}},{{alarm_description}},{{temperature}},{{humidity}},{{pressure}},{{wind_speed}},{{lux}},{{uv}},{{rain_daily}}';
        
        channels.push({ 
            type: 'csv',
            csv: {
                path: csvPath,
                max_days: csvMaxDays,
                message: csvMessage
            }
        });
    }
    
    if (document.getElementById('deliveryJSON').checked) {
        const jsonPath = document.getElementById('jsonPath').value;
        const jsonMaxDays = parseInt(document.getElementById('jsonMaxDays').value) || 30;
        const jsonMessage = document.getElementById('jsonMessage').value || '{"timestamp": "{{timestamp}}", "message": "ALARM: {{alarm_name}} triggered", "alarm": {{alarm_info}}, "sensors": {{sensor_info}}}';
        
        channels.push({ 
            type: 'json',
            json: {
                path: jsonPath,
                max_days: jsonMaxDays,
                message: jsonMessage
            }
        });
    }
    
    const alarmData = {
        name: document.getElementById('alarmName').value,
        description: document.getElementById('alarmDescription').value,
        condition: document.getElementById('alarmCondition').value,
        tags: selectedTags,
        cooldown: parseInt(document.getElementById('alarmCooldown').value),
        enabled: document.getElementById('alarmEnabled').checked,
        channels: channels
    };
    
    // Track original name for updates (in case name changed)
    const originalName = currentAlarm ? currentAlarm.name : null;
    const endpoint = currentAlarm ? `/api/alarms/update?oldName=${encodeURIComponent(originalName)}` : '/api/alarms/create';
    
    try {
        const response = await fetch(endpoint, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(alarmData)
        });
        
        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }
        
        showNotification(currentAlarm ? 'Alarm updated successfully' : 'Alarm created successfully', 'success');
        closeModal();
        await loadAlarms();
        await loadTags();
    } catch (error) {
        showNotification('Error: ' + error.message, 'error');
    }
}

async function deleteAlarm(name) {
    if (!confirm('Are you sure you want to delete alarm "' + name + '"?')) return;
    
    try {
        const response = await fetch('/api/alarms/delete?name=' + encodeURIComponent(name), {
            method: 'POST'
        });
        
        if (!response.ok) {
            throw new Error(await response.text());
        }
        
        // If the deleted alarm is currently being viewed/edited, close the modal
        if (currentAlarm && currentAlarm.name === name) {
            closeModal();
            closeJSONModal();
            currentAlarm = null;
        }
        
        // Remove the alarm from the local array immediately
        alarms = alarms.filter(a => a.name !== name);
        
        // Re-render the UI immediately
        renderAlarms();
        
        showNotification('Alarm deleted successfully', 'success');
        
        // Reload from server to ensure consistency and update tags
        await loadAlarms();
        await loadTags();
    } catch (error) {
        showNotification('Error: ' + error.message, 'error');
    }
}

async function saveAll() {
    const response = await fetch('/api/config');
    const config = await response.json();
    
    const blob = new Blob([JSON.stringify(config, null, 2)], {type: 'application/json'});
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'alarms.json';
    a.click();
    
    showNotification('Configuration saved', 'success');
}

function showNotification(message, type) {
    const notification = document.getElementById('notification');
    notification.textContent = message;
    notification.className = 'notification ' + type + ' show';
    setTimeout(() => {
        notification.classList.remove('show');
    }, 3000);
}

function updateLastUpdateTimestamp() {
    const lastUpdateElement = document.getElementById('last-update');
    if (lastUpdateElement) {
        const now = new Date();
        const lastUpdateText = now.toLocaleString('en-US', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            hour12: false
        });
        lastUpdateElement.textContent = lastUpdateText;
    }
}

// ============================================
// Emoji Picker Functions
// ============================================

let currentEmojiTarget = null;

function showEmojiPicker(targetId) {
    currentEmojiTarget = targetId;
    document.getElementById('emojiModal').classList.add('active');
}

function closeEmojiModal() {
    document.getElementById('emojiModal').classList.remove('active');
    currentEmojiTarget = null;
}

function insertEmoji(targetId, emoji) {
    const textarea = document.getElementById(targetId);
    if (textarea) {
        const start = textarea.selectionStart;
        const end = textarea.selectionEnd;
        const text = textarea.value;
        const before = text.substring(0, start);
        const after = text.substring(end, text.length);
        textarea.value = before + emoji + after;
        textarea.selectionStart = textarea.selectionEnd = start + emoji.length;
        textarea.focus();
    }
    closeEmojiModal();
}

// Update emoji buttons to use the current target
document.addEventListener('DOMContentLoaded', function() {
    // Update emoji modal buttons to use current target
    const emojiModal = document.getElementById('emojiModal');
    if (emojiModal) {
        const emojiButtons = emojiModal.querySelectorAll('button[onclick*="insertEmoji"]');
        emojiButtons.forEach(button => {
            const onclick = button.getAttribute('onclick');
            button.setAttribute('onclick', onclick.replace('consoleMessage', currentEmojiTarget || 'consoleMessage'));
        });
    }
});

init();
