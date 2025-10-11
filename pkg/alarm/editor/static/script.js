let alarms = [];
let currentAlarm = null;
let allTags = [];
let selectedTags = [];

async function init() {
    await loadAlarms();
    await loadTags();
    document.getElementById('searchName').addEventListener('input', filterAlarms);
    document.getElementById('filterTag').addEventListener('change', filterAlarms);
    document.getElementById('alarmForm').addEventListener('submit', handleSubmit);
    initTagSelector();
}

async function loadAlarms() {
    const response = await fetch('/api/config');
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
    // Show/hide custom message sections based on selected delivery methods
    const consoleChecked = document.getElementById('deliveryConsole').checked;
    const syslogChecked = document.getElementById('deliverySyslog').checked;
    const oslogChecked = document.getElementById('deliveryOslog').checked;
    const eventlogChecked = document.getElementById('deliveryEventlog').checked;
    const emailChecked = document.getElementById('deliveryEmail').checked;
    const smsChecked = document.getElementById('deliverySMS').checked;
    
    document.getElementById('consoleMessageSection').style.display = consoleChecked ? 'block' : 'none';
    document.getElementById('syslogMessageSection').style.display = syslogChecked ? 'block' : 'none';
    document.getElementById('oslogMessageSection').style.display = oslogChecked ? 'block' : 'none';
    document.getElementById('eventlogMessageSection').style.display = eventlogChecked ? 'block' : 'none';
    document.getElementById('emailMessageSection').style.display = emailChecked ? 'block' : 'none';
    document.getElementById('smsMessageSection').style.display = smsChecked ? 'block' : 'none';
}

function toggleCustomMessages() {
    const useDefault = document.getElementById('useDefaultMessage').checked;
    document.getElementById('defaultMessageSection').style.display = useDefault ? 'block' : 'none';
    document.getElementById('customMessageSections').style.display = useDefault ? 'none' : 'block';
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
    
    // Reset messages
    document.getElementById('useDefaultMessage').checked = true;
    document.getElementById('defaultMessage').value = '🚨 ALARM: {{alarm_name}}\nStation: {{station}}\nTime: {{timestamp}}';
    document.getElementById('consoleMessage').value = '';
    document.getElementById('syslogMessage').value = '';
    document.getElementById('oslogMessage').value = '';
    document.getElementById('eventlogMessage').value = '';
    document.getElementById('emailTo').value = '';
    document.getElementById('emailSubject').value = '';
    document.getElementById('emailBody').value = '';
    document.getElementById('smsTo').value = '';
    document.getElementById('smsMessage').value = '';
    
    selectedTags = [];
    renderSelectedTags();
    document.getElementById('tagSearchInput').value = '';
    
    toggleMessageSections();
    toggleCustomMessages();
    
    document.getElementById('editModal').classList.add('active');
}

function editAlarm(name) {
    currentAlarm = alarms.find(a => a.name === name);
    if (!currentAlarm) return;
    
    document.getElementById('alarmName').value = currentAlarm.name;
    document.getElementById('alarmName').readOnly = true;
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
    
    // Load messages from channels
    let hasCustomMessages = false;
    let defaultMsg = '🚨 ALARM: {{alarm_name}}\nStation: {{station}}\nTime: {{timestamp}}';
    
    channels.forEach(channel => {
        if (channel.type === 'console' && channel.template) {
            document.getElementById('consoleMessage').value = channel.template;
            defaultMsg = channel.template;
        } else if (channel.type === 'syslog' && channel.template) {
            document.getElementById('syslogMessage').value = channel.template;
            if (!hasCustomMessages && channel.template !== defaultMsg) hasCustomMessages = true;
        } else if (channel.type === 'oslog' && channel.template) {
            document.getElementById('oslogMessage').value = channel.template;
            if (!hasCustomMessages && channel.template !== defaultMsg) hasCustomMessages = true;
        } else if (channel.type === 'eventlog' && channel.template) {
            document.getElementById('eventlogMessage').value = channel.template;
            if (!hasCustomMessages && channel.template !== defaultMsg) hasCustomMessages = true;
        } else if (channel.type === 'email' && channel.email) {
            document.getElementById('emailTo').value = (channel.email.to || []).join(', ');
            document.getElementById('emailSubject').value = channel.email.subject || '';
            document.getElementById('emailBody').value = channel.email.body || '';
            hasCustomMessages = true;
        } else if (channel.type === 'sms' && channel.sms) {
            document.getElementById('smsTo').value = (channel.sms.to || []).join(', ');
            document.getElementById('smsMessage').value = channel.sms.message || '';
            hasCustomMessages = true;
        }
    });
    
    document.getElementById('defaultMessage').value = defaultMsg;
    document.getElementById('useDefaultMessage').checked = !hasCustomMessages;
    
    toggleMessageSections();
    toggleCustomMessages();
    
    document.getElementById('editModal').classList.add('active');
}

function closeModal() {
    document.getElementById('editModal').classList.remove('active');
}

function closeJSONModal() {
    document.getElementById('jsonModal').classList.remove('active');
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
    const jsonString = JSON.stringify(data, null, 2);
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
    
    // Build channels array from selected delivery methods
    const channels = [];
    const useDefault = document.getElementById('useDefaultMessage').checked;
    const defaultMessage = document.getElementById('defaultMessage').value || '🚨 ALARM: {{alarm_name}}\nStation: {{station}}\nTime: {{timestamp}}';
    
    if (document.getElementById('deliveryConsole').checked) {
        const template = useDefault ? defaultMessage : (document.getElementById('consoleMessage').value || defaultMessage);
        channels.push({ 
            type: 'console',
            template: template
        });
    }
    if (document.getElementById('deliverySyslog').checked) {
        const template = useDefault ? defaultMessage : (document.getElementById('syslogMessage').value || defaultMessage);
        channels.push({ 
            type: 'syslog',
            template: template
        });
    }
    if (document.getElementById('deliveryOslog').checked) {
        const template = useDefault ? defaultMessage : (document.getElementById('oslogMessage').value || defaultMessage);
        channels.push({ 
            type: 'oslog',
            template: template
        });
    }
    if (document.getElementById('deliveryEventlog').checked) {
        const template = useDefault ? defaultMessage : (document.getElementById('eventlogMessage').value || defaultMessage);
        channels.push({ 
            type: 'eventlog',
            template: template
        });
    }
    if (document.getElementById('deliveryEmail').checked) {
        const emailTo = document.getElementById('emailTo').value;
        const emailSubject = document.getElementById('emailSubject').value;
        const emailBody = document.getElementById('emailBody').value;
        
        channels.push({ 
            type: 'email',
            email: {
                to: emailTo ? emailTo.split(',').map(e => e.trim()).filter(e => e) : ['admin@example.com'],
                subject: emailSubject || '⚠️ Weather Alarm: {{alarm_name}}',
                body: emailBody || defaultMessage
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
    
    const alarmData = {
        name: document.getElementById('alarmName').value,
        description: document.getElementById('alarmDescription').value,
        condition: document.getElementById('alarmCondition').value,
        tags: selectedTags,
        cooldown: parseInt(document.getElementById('alarmCooldown').value),
        enabled: document.getElementById('alarmEnabled').checked,
        channels: channels
    };
    
    const endpoint = currentAlarm ? '/api/alarms/update' : '/api/alarms/create';
    
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
        
        showNotification('Alarm deleted successfully', 'success');
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

init();
