package editor

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Alarm Configuration Editor</title>
    <link rel="stylesheet" href="/alarm-editor/static/styles.css">
    <link rel="stylesheet" href="/alarm-editor/static/themes.css">
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>⚡ Tempest Alarm Editor</h1>
            <p>Create and manage weather alarms with real-time monitoring</p>
            <div class="config-path-display">
                <span class="label">📁 Watching:</span>
                <span class="path">{{.ConfigPath}}</span>
                <span class="label" style="margin-left: 20px;">🕒 Last read:</span>
                <span class="path">{{.LastLoad}}</span>
            </div>
        </div>
        
        <div class="toolbar">
            <input type="text" id="searchName" placeholder="🔍 Search by name..." />
            <select id="filterTag">
                <option value="">All Tags</option>
            </select>
            <button class="btn btn-primary" onclick="showCreateModal()">+ New Alarm</button>
            <button class="btn btn-info" onclick="showFullJSON()">📄 View Full JSON</button>
            <button class="btn btn-warning" onclick="showEditContactsModal()">👥 Edit Contacts</button>
            <button class="btn btn-warning" onclick="showEditTagsModal()">🏷️ Edit Tags</button>
            <button class="btn btn-success" onclick="saveAll()">💾 Save All</button>
        </div>
        
        <div class="content">
            <div id="alarmList" class="alarm-grid"></div>
            <div id="emptyState" class="empty-state" style="display:none;">
                <h3>No alarms found</h3>
                <p>Create your first alarm to get started</p>
            </div>
        </div>
    </div>
    
    <div id="jsonModal" class="modal">
        <div class="modal-content wide">
            <div class="modal-header" id="jsonModalTitle">JSON View</div>
            <div class="json-viewer" id="jsonContent"></div>
            <div class="modal-actions">
                <button type="button" class="btn btn-secondary" onclick="closeJSONModal()">Close</button>
                <button type="button" class="btn btn-primary" onclick="copyJSON()">📋 Copy to Clipboard</button>
            </div>
        </div>
    </div>
    
    <div id="editModal" class="modal">
        <div class="modal-content">
            <div class="modal-header">Edit Alarm</div>
            <form id="alarmForm">
                <div class="form-group">
                    <label>Name *</label>
                    <input type="text" id="alarmName" required />
                </div>
                
                <div class="form-group">
                    <label>Description</label>
                    <input type="text" id="alarmDescription" />
                </div>
                
                <div class="form-group">
                    <label>Condition *</label>
                    <div class="sensor-fields">
                        <button type="button" class="sensor-field-btn" onclick="insertField('humidity')">humidity</button>
                        <button type="button" class="sensor-field-btn" onclick="insertField('lightning_count')">lightning_count</button>
                        <button type="button" class="sensor-field-btn" onclick="insertField('lightning_distance')">lightning_distance</button>
                        <button type="button" class="sensor-field-btn" onclick="insertField('lux')">lux</button>
                        <button type="button" class="sensor-field-btn" onclick="insertField('pressure')">pressure</button>
                        <button type="button" class="sensor-field-btn" onclick="insertField('rain_daily')">rain_daily</button>
                        <button type="button" class="sensor-field-btn" onclick="insertField('rain_rate')">rain_rate</button>
                        <button type="button" class="sensor-field-btn" onclick="insertField('temperature')">temperature</button>
                        <button type="button" class="sensor-field-btn" onclick="insertField('uv')">uv</button>
                        <button type="button" class="sensor-field-btn" onclick="insertField('wind_direction')">wind_direction</button>
                        <button type="button" class="sensor-field-btn" onclick="insertField('wind_gust')">wind_gust</button>
                        <button type="button" class="sensor-field-btn" onclick="insertField('wind_speed')">wind_speed</button>
                    </div>
                    <textarea id="alarmCondition" required></textarea>
                    <button type="button" class="btn btn-info" onclick="validateCondition()" style="margin-top: 8px;">✓ Validate Condition</button>
                    <div id="validationResult" style="margin-top: 8px; padding: 8px; border-radius: 4px; display: none;"></div>
                    <small>Click sensor names above to insert into condition. Supports units: 80F or 26.7C (temp), 25mph or 11.2m/s (wind). Change detection: *field (any change), &gt;field (increase), &lt;field (decrease). Examples: temperature &gt; 85F, *lightning_count (any strike), &gt;rain_rate (rain increasing), &lt;lightning_distance (lightning closer)</small>
                </div>
                
                <div class="form-group">
                    <label>Delivery Methods *</label>
                    <div class="delivery-methods">
                        <label class="delivery-method">
                            <input type="checkbox" id="deliveryConsole" checked onchange="toggleMessageSections()" />
                            <span>📟 Console</span>
                        </label>
                        <label class="delivery-method">
                            <input type="checkbox" id="deliverySyslog" onchange="toggleMessageSections()" />
                            <span>📋 Syslog</span>
                        </label>
                        <label class="delivery-method">
                            <input type="checkbox" id="deliveryOslog" onchange="toggleMessageSections()" />
                            <span>🍎 OSLog</span>
                        </label>
                        <label class="delivery-method">
                            <input type="checkbox" id="deliveryEventlog" onchange="toggleMessageSections()" />
                            <span>📊 Event Log</span>
                        </label>
                        <label class="delivery-method">
                            <input type="checkbox" id="deliveryEmail" onchange="toggleMessageSections()" />
                            <span>✉️ Email</span>
                        </label>
                        <label class="delivery-method">
                            <input type="checkbox" id="deliverySMS" onchange="toggleMessageSections()" />
                            <span>📱 SMS</span>
                        </label>
                        <label class="delivery-method">
                            <input type="checkbox" id="deliveryWebhook" onchange="toggleMessageSections()" />
                            <span>🌐 Webhook</span>
                        </label>
                        <label class="delivery-method">
                            <input type="checkbox" id="deliveryCSV" onchange="toggleMessageSections()" />
                            <span>📊 CSV File</span>
                        </label>
                        <label class="delivery-method">
                            <input type="checkbox" id="deliveryJSON" onchange="toggleMessageSections()" />
                            <span>📄 JSON File</span>
                        </label>
                    </div>
                    <small>Select at least one delivery method. Each method will show its configuration below with defaults pre-populated.</small>
                </div>
                
                <div id="messageSections">
                    <div id="consoleMessageSection" class="form-group message-input-section" style="display:block;">
                        <div class="message-header">
                            <label>📟 Console Message</label>
                            <div style="display: flex; gap: 8px; align-items: center;">
                                <select onchange="insertVariable('consoleMessage')" class="variable-dropdown">
                                    <option value="">📋 Insert Variable...</option>
                                    <option value="{{ "{{" }}app_info}}">{{ "{{" }}app_info}} - Application info (version, uptime)</option>
                                    <option value="{{ "{{" }}alarm_info}}">{{ "{{" }}alarm_info}} - Alarm info (name, desc, condition)</option>
                                    <option value="{{ "{{" }}sensor_info}}">{{ "{{" }}sensor_info}} - Sensor values that triggered alarm</option>
                                    <option value="{{ "{{" }}alarm_name}}">{{ "{{" }}alarm_name}} - Alarm name</option>
                                    <option value="{{ "{{" }}alarm_description}}">{{ "{{" }}alarm_description}} - Alarm description</option>
                                    <option value="{{ "{{" }}alarm_condition}}">{{ "{{" }}alarm_condition}} - Alarm condition</option>
                                    <option value="{{ "{{" }}station}}">{{ "{{" }}station}} - Station name</option>
                                    <option value="{{ "{{" }}timestamp}}">{{ "{{" }}timestamp}} - Current time</option>
                                    <option value="{{ "{{" }}temperature}}">{{ "{{" }}temperature}} - Temperature °C (current)</option>
                                    <option value="{{ "{{" }}temperature_f}}">{{ "{{" }}temperature_f}} - Temperature °F (current)</option>
                                    <option value="{{ "{{" }}humidity}}">{{ "{{" }}humidity}} - Humidity % (current)</option>
                                    <option value="{{ "{{" }}pressure}}">{{ "{{" }}pressure}} - Pressure mb (current)</option>
                                    <option value="{{ "{{" }}wind_speed}}">{{ "{{" }}wind_speed}} - Wind Speed m/s (current)</option>
                                    <option value="{{ "{{" }}wind_gust}}">{{ "{{" }}wind_gust}} - Wind Gust m/s (current)</option>
                                    <option value="{{ "{{" }}wind_direction}}">{{ "{{" }}wind_direction}} - Wind Direction° (current)</option>
                                    <option value="{{ "{{" }}lux}}">{{ "{{" }}lux}} - Light Lux (current)</option>
                                    <option value="{{ "{{" }}uv}}">{{ "{{" }}uv}} - UV Index (current)</option>
                                    <option value="{{ "{{" }}rain_rate}}">{{ "{{" }}rain_rate}} - Rain Rate mm (current)</option>
                                    <option value="{{ "{{" }}rain_daily}}">{{ "{{" }}rain_daily}} - Daily Rain mm (current)</option>
                                    <option value="{{ "{{" }}lightning_count}}">{{ "{{" }}lightning_count}} - Lightning Strikes (current)</option>
                                    <option value="{{ "{{" }}lightning_distance}}">{{ "{{" }}lightning_distance}} - Lightning Distance km (current)</option>
                                </select>
                                <button type="button" class="btn btn-secondary" onclick="showEmojiPicker('consoleMessage')" title="Insert Emoji">😀</button>
                            </div>
                        </div>
                        <textarea id="consoleMessage" rows="5" placeholder="Console-specific message..."></textarea>
                    </div>
                    
                    <div id="syslogMessageSection" class="form-group message-input-section" style="display:none;">
                        <div class="message-header">
                            <label>📋 Syslog Message</label>
                            <div style="display: flex; gap: 8px; align-items: center;">
                                <select onchange="insertVariable('syslogMessage')" class="variable-dropdown">
                                    <option value="">📋 Insert Variable...</option>
                                    <option value="{{ "{{" }}app_info}}">{{ "{{" }}app_info}} - Application info (version, uptime)</option>
                                    <option value="{{ "{{" }}alarm_info}}">{{ "{{" }}alarm_info}} - Alarm info (name, desc, condition)</option>
                                    <option value="{{ "{{" }}sensor_info}}">{{ "{{" }}sensor_info}} - Sensor values that triggered alarm</option>
                                    <option value="{{ "{{" }}alarm_name}}">{{ "{{" }}alarm_name}} - Alarm name</option>
                                    <option value="{{ "{{" }}alarm_description}}">{{ "{{" }}alarm_description}} - Alarm description</option>
                                    <option value="{{ "{{" }}alarm_condition}}">{{ "{{" }}alarm_condition}} - Alarm condition</option>
                                    <option value="{{ "{{" }}station}}">{{ "{{" }}station}} - Station name</option>
                                    <option value="{{ "{{" }}timestamp}}">{{ "{{" }}timestamp}} - Current time</option>
                                    <option value="{{ "{{" }}temperature}}">{{ "{{" }}temperature}} - Temperature °C (current)</option>
                                    <option value="{{ "{{" }}temperature_f}}">{{ "{{" }}temperature_f}} - Temperature °F (current)</option>
                                    <option value="{{ "{{" }}humidity}}">{{ "{{" }}humidity}} - Humidity % (current)</option>
                                    <option value="{{ "{{" }}pressure}}">{{ "{{" }}pressure}} - Pressure mb (current)</option>
                                    <option value="{{ "{{" }}wind_speed}}">{{ "{{" }}wind_speed}} - Wind Speed m/s (current)</option>
                                </select>
                                <button type="button" class="btn btn-secondary" onclick="showEmojiPicker('syslogMessage')" title="Insert Emoji">😀</button>
                            </div>
                        </div>
                        <textarea id="syslogMessage" rows="5" placeholder="Syslog-specific message..."></textarea>
                    </div>
                    
                    <div id="oslogMessageSection" class="form-group message-input-section" style="display:none;">
                        <div class="message-header">
                            <label>🍎 OSLog Message (macOS Unified Logging)</label>
                            <div style="display: flex; gap: 8px; align-items: center;">
                                <select onchange="insertVariable('oslogMessage')" class="variable-dropdown">
                                    <option value="">📋 Insert Variable...</option>
                                    <option value="{{ "{{" }}app_info}}">{{ "{{" }}app_info}} - Application info (version, uptime)</option>
                                    <option value="{{ "{{" }}alarm_info}}">{{ "{{" }}alarm_info}} - Alarm info (name, desc, condition)</option>
                                    <option value="{{ "{{" }}sensor_info}}">{{ "{{" }}sensor_info}} - Sensor values that triggered alarm</option>
                                    <option value="{{ "{{" }}alarm_name}}">{{ "{{" }}alarm_name}} - Alarm name</option>
                                    <option value="{{ "{{" }}alarm_description}}">{{ "{{" }}alarm_description}} - Alarm description</option>
                                </select>
                                <button type="button" class="btn btn-secondary" onclick="showEmojiPicker('oslogMessage')" title="Insert Emoji">😀</button>
                            </div>
                        </div>
                        <textarea id="oslogMessage" rows="5" placeholder="OSLog-specific message (visible in Console.app and log command)..."></textarea>
                    </div>
                    
                    <div id="eventlogMessageSection" class="form-group message-input-section" style="display:none;">
                        <div class="message-header">
                            <label>📊 Event Log Message</label>
                            <div style="display: flex; gap: 8px; align-items: center;">
                                <select onchange="insertVariable('eventlogMessage')" class="variable-dropdown">
                                    <option value="">📋 Insert Variable...</option>
                                    <option value="{{ "{{" }}app_info}}">{{ "{{" }}app_info}} - Application info (version, uptime)</option>
                                    <option value="{{ "{{" }}alarm_info}}">{{ "{{" }}alarm_info}} - Alarm info (name, desc, condition)</option>
                                    <option value="{{ "{{" }}sensor_info}}">{{ "{{" }}sensor_info}} - Sensor values that triggered alarm</option>
                                    <option value="{{ "{{" }}alarm_name}}">{{ "{{" }}alarm_name}} - Alarm name</option>
                                    <option value="{{ "{{" }}alarm_description}}">{{ "{{" }}alarm_description}} - Alarm description</option>
                                </select>
                                <button type="button" class="btn btn-secondary" onclick="showEmojiPicker('eventlogMessage')" title="Insert Emoji">😀</button>
                            </div>
                        </div>
                        <textarea id="eventlogMessage" rows="5" placeholder="Event log-specific message..."></textarea>
                    </div>
                    
                    <div id="emailMessageSection" class="form-group message-input-section" style="display:none;">
                        <div class="message-header">
                            <label>✉️ Email Configuration</label>
                            <div style="display: flex; gap: 8px; align-items: center;">
                                <select onchange="insertVariable('emailBody')" class="variable-dropdown">
                                    <option value="">📋 Insert Variable...</option>
                                    <option value="{{ "{{" }}app_info}}">{{ "{{" }}app_info}} - Application info (version, uptime)</option>
                                    <option value="{{ "{{" }}alarm_info}}">{{ "{{" }}alarm_info}} - Alarm info (name, desc, condition)</option>
                                    <option value="{{ "{{" }}sensor_info}}">{{ "{{" }}sensor_info}} - Sensor values that triggered alarm</option>
                                    <option value="{{ "{{" }}alarm_name}}">{{ "{{" }}alarm_name}} - Alarm name</option>
                                    <option value="{{ "{{" }}alarm_description}}">{{ "{{" }}alarm_description}} - Alarm description</option>
                                    <option value="{{ "{{" }}station}}">{{ "{{" }}station}} - Station name</option>
                                    <option value="{{ "{{" }}timestamp}}">{{ "{{" }}timestamp}} - Current time</option>
                                </select>
                                <button type="button" class="btn btn-secondary" onclick="showEmojiPicker('emailBody')" title="Insert Emoji">😀</button>
                            </div>
                        </div>
                        <label for="emailTo" style="font-weight: 600;">To: <span style="color: red;">*</span></label>
                        <div class="contact-selector-container">
                            <div class="selected-contacts" id="selectedEmailContacts"></div>
                            <div class="contact-dropdown-wrapper">
                                <input type="text" 
                                       id="emailContactSearch" 
                                       class="contact-search-input" 
                                       placeholder="Search contacts or enter email..." 
                                       autocomplete="off" />
                                <select id="emailContactSelect" onchange="addContactEmail()" style="display: none;">
                                    <option value="">📋 Add Contact...</option>
                                </select>
                                <div id="emailContactDropdown" class="contact-dropdown"></div>
                            </div>
                        </div>
                        <label for="emailSubject" style="margin-top: 10px; font-weight: 600;">Subject:</label>
                        <input type="text" id="emailSubject" placeholder="Tempest Alert: {{ "{{" }}alarm_name{{ "}}" }}" />
                        <label style="margin-top: 10px;">
                            <input type="checkbox" id="emailHtml" />
                            Send as HTML formatted email
                        </label>
                        <label for="emailBody" style="margin-top: 10px; font-weight: 600;">Body:</label>
                        <textarea id="emailBody" rows="8" placeholder="Email body..."></textarea>
                        <small>When HTML is enabled, use HTML tags like &lt;h1&gt;, &lt;p&gt;, &lt;strong&gt;, &lt;br&gt;, etc. for formatting</small>
                    </div>
                    
                    <div id="smsMessageSection" class="form-group message-input-section" style="display:none;">
                        <div class="message-header">
                            <label>📱 SMS Configuration</label>
                            <div style="display: flex; gap: 8px; align-items: center;">
                                <select onchange="insertVariable('smsMessage')" class="variable-dropdown">
                                    <option value="">📋 Insert Variable...</option>
                                    <option value="{{ "{{" }}app_info}}">{{ "{{" }}app_info}} - Application info (version, uptime)</option>
                                    <option value="{{ "{{" }}alarm_info}}">{{ "{{" }}alarm_info}} - Alarm info (name, desc, condition)</option>
                                    <option value="{{ "{{" }}sensor_info}}">{{ "{{" }}sensor_info}} - Sensor values that triggered alarm</option>
                                    <option value="{{ "{{" }}alarm_name}}">{{ "{{" }}alarm_name}} - Alarm name</option>
                                    <option value="{{ "{{" }}station}}">{{ "{{" }}station}} - Station name</option>
                                    <option value="{{ "{{" }}timestamp}}">{{ "{{" }}timestamp}} - Current time</option>
                                </select>
                                <button type="button" class="btn btn-secondary" onclick="showEmojiPicker('smsMessage')" title="Insert Emoji">😀</button>
                            </div>
                        </div>
                        <label for="smsTo" style="font-weight: 600;">To: <span style="color: red;">*</span></label>
                        <div class="contact-selector-container">
                            <div class="selected-contacts" id="selectedSMSContacts"></div>
                            <div class="contact-dropdown-wrapper">
                                <input type="text" 
                                       id="smsContactSearch" 
                                       class="contact-search-input" 
                                       placeholder="Search contacts or enter phone number..." 
                                       autocomplete="off" />
                                <select id="smsContactSelect" onchange="addContactSMS()" style="display: none;">
                                    <option value="">📋 Add Contact...</option>
                                </select>
                                <div id="smsContactDropdown" class="contact-dropdown"></div>
                            </div>
                        </div>
                        <label for="smsMessage" style="margin-top: 10px; font-weight: 600;">Message:</label>
                        <textarea id="smsMessage" rows="3" placeholder="SMS message (keep it short)..."></textarea>
                    </div>
                    
                    <div id="webhookMessageSection" class="form-group message-input-section" style="display:none;">
                        <div class="message-header">
                            <label>🌐 Webhook Configuration</label>
                            <div style="display: flex; gap: 8px; align-items: center;">
                                <select onchange="insertVariable('webhookBody')" class="variable-dropdown">
                                    <option value="">📋 Insert Variable...</option>
                                    <option value="{{ "{{" }}app_info}}">{{ "{{" }}app_info}} - Application info (version, uptime)</option>
                                    <option value="{{ "{{" }}alarm_info}}">{{ "{{" }}alarm_info}} - Alarm info (name, desc, condition)</option>
                                    <option value="{{ "{{" }}sensor_info}}">{{ "{{" }}sensor_info}} - Sensor values that triggered alarm</option>
                                    <option value="{{ "{{" }}alarm_name}}">{{ "{{" }}alarm_name}} - Alarm name</option>
                                    <option value="{{ "{{" }}alarm_description}}">{{ "{{" }}alarm_description}} - Alarm description</option>
                                    <option value="{{ "{{" }}alarm_condition}}">{{ "{{" }}alarm_condition}} - Alarm condition</option>
                                    <option value="{{ "{{" }}station}}">{{ "{{" }}station}} - Station name</option>
                                    <option value="{{ "{{" }}timestamp}}">{{ "{{" }}timestamp}} - Current time</option>
                                    <option value="{{ "{{" }}temperature}}">{{ "{{" }}temperature}} - Temperature °C (current)</option>
                                    <option value="{{ "{{" }}temperature_f}}">{{ "{{" }}temperature_f}} - Temperature °F (current)</option>
                                    <option value="{{ "{{" }}humidity}}">{{ "{{" }}humidity}} - Humidity % (current)</option>
                                    <option value="{{ "{{" }}pressure}}">{{ "{{" }}pressure}} - Pressure mb (current)</option>
                                    <option value="{{ "{{" }}wind_speed}}">{{ "{{" }}wind_speed}} - Wind Speed m/s (current)</option>
                                    <option value="{{ "{{" }}wind_gust}}">{{ "{{" }}wind_gust}} - Wind Gust m/s (current)</option>
                                    <option value="{{ "{{" }}wind_direction}}">{{ "{{" }}wind_direction}} - Wind Direction° (current)</option>
                                    <option value="{{ "{{" }}lux}}">{{ "{{" }}lux}} - Light Lux (current)</option>
                                    <option value="{{ "{{" }}uv}}">{{ "{{" }}uv}} - UV Index (current)</option>
                                    <option value="{{ "{{" }}rain_rate}}">{{ "{{" }}rain_rate}} - Rain Rate mm (current)</option>
                                    <option value="{{ "{{" }}rain_daily}}">{{ "{{" }}rain_daily}} - Daily Rain mm (current)</option>
                                    <option value="{{ "{{" }}lightning_count}}">{{ "{{" }}lightning_count}} - Lightning Strikes (current)</option>
                                    <option value="{{ "{{" }}lightning_distance}}">{{ "{{" }}lightning_distance}} - Lightning Distance km (current)</option>
                                </select>
                                <button type="button" class="btn btn-secondary" onclick="showEmojiPicker('webhookBody')" title="Insert Emoji">😀</button>
                            </div>
                        </div>
                        <label for="webhookUrl" style="font-weight: 600;">URL: <span style="color: red;">*</span></label>
                        <input type="url" id="webhookUrl" placeholder="https://api.example.com/webhooks/alert" />
                        <label for="webhookMethod" style="margin-top: 10px; font-weight: 600;">Method:</label>
                        <select id="webhookMethod">
                            <option value="POST">POST</option>
                            <option value="PUT">PUT</option>
                            <option value="PATCH">PATCH</option>
                        </select>
                        <label for="webhookHeaders" style="margin-top: 10px; font-weight: 600;">Headers (JSON):</label>
                        <textarea id="webhookHeaders" rows="3" placeholder='{"Authorization": "Bearer token", "Content-Type": "application/json"}'></textarea>
                        <label for="webhookBody" style="margin-top: 10px; font-weight: 600;">Body:</label>
                        <textarea id="webhookBody" rows="8" placeholder="Webhook body (JSON or plain text)..."></textarea>
                        <label for="webhookContentType" style="margin-top: 10px; font-weight: 600;">Content Type:</label>
                        <input type="text" id="webhookContentType" value="application/json" placeholder="application/json" />
                        <small>Headers should be valid JSON. Body supports template variables like &#123;&#123;alarm_name&#125;&#125;. Content type defaults to application/json.</small>
                    </div>
                    
                    <div id="csvMessageSection" class="form-group message-input-section" style="display:none;">
                        <div class="message-header">
                            <label>📊 CSV File Configuration</label>
                            <div style="display: flex; gap: 8px; align-items: center;">
                                <select onchange="insertVariable('csvMessage')" class="variable-dropdown">
                                    <option value="">📋 Insert Variable...</option>
                                    <option value="{{ "{{" }}app_info}}">{{ "{{" }}app_info}} - Application info (version, uptime)</option>
                                    <option value="{{ "{{" }}alarm_info}}">{{ "{{" }}alarm_info}} - Alarm info (name, desc, condition)</option>
                                    <option value="{{ "{{" }}sensor_info}}">{{ "{{" }}sensor_info}} - Sensor values that triggered alarm</option>
                                    <option value="{{ "{{" }}alarm_name}}">{{ "{{" }}alarm_name}} - Alarm name</option>
                                    <option value="{{ "{{" }}alarm_description}}">{{ "{{" }}alarm_description}} - Alarm description</option>
                                    <option value="{{ "{{" }}alarm_condition}}">{{ "{{" }}alarm_condition}} - Alarm condition</option>
                                    <option value="{{ "{{" }}station}}">{{ "{{" }}station}} - Station name</option>
                                    <option value="{{ "{{" }}timestamp}}">{{ "{{" }}timestamp}} - Current time</option>
                                    <option value="{{ "{{" }}temperature}}">{{ "{{" }}temperature}} - Temperature °C (current)</option>
                                    <option value="{{ "{{" }}temperature_f}}">{{ "{{" }}temperature_f}} - Temperature °F (current)</option>
                                    <option value="{{ "{{" }}humidity}}">{{ "{{" }}humidity}} - Humidity % (current)</option>
                                    <option value="{{ "{{" }}pressure}}">{{ "{{" }}pressure}} - Pressure mb (current)</option>
                                    <option value="{{ "{{" }}wind_speed}}">{{ "{{" }}wind_speed}} - Wind Speed m/s (current)</option>
                                    <option value="{{ "{{" }}wind_gust}}">{{ "{{" }}wind_gust}} - Wind Gust m/s (current)</option>
                                    <option value="{{ "{{" }}wind_direction}}">{{ "{{" }}wind_direction}} - Wind Direction° (current)</option>
                                    <option value="{{ "{{" }}lux}}">{{ "{{" }}lux}} - Light Lux (current)</option>
                                    <option value="{{ "{{" }}uv}}">{{ "{{" }}uv}} - UV Index (current)</option>
                                    <option value="{{ "{{" }}rain_rate}}">{{ "{{" }}rain_rate}} - Rain Rate mm (current)</option>
                                    <option value="{{ "{{" }}rain_daily}}">{{ "{{" }}rain_daily}} - Daily Rain mm (current)</option>
                                    <option value="{{ "{{" }}lightning_count}}">{{ "{{" }}lightning_count}} - Lightning Strikes (current)</option>
                                    <option value="{{ "{{" }}lightning_distance}}">{{ "{{" }}lightning_distance}} - Lightning Distance km (current)</option>
                                </select>
                                <button type="button" class="btn btn-secondary" onclick="showEmojiPicker('csvMessage')" title="Insert Emoji">😀</button>
                            </div>
                        </div>
                        <label for="csvPath" style="font-weight: 600;">File Path: <span style="color: red;">*</span></label>
                        <input type="text" id="csvPath" placeholder="/tmp/tempest-alarms.csv" />
                        <label for="csvMaxDays" style="margin-top: 10px; font-weight: 600;">Max Days (0 = unlimited):</label>
                        <input type="number" id="csvMaxDays" value="30" min="0" placeholder="30" />
                        <label for="csvMessage" style="margin-top: 10px; font-weight: 600;">Message Template: <span style="color: red;">*</span></label>
                        <textarea id="csvMessage" rows="3" placeholder="CSV message template..."></textarea>
                        <small>CSV files will be rotated when max days is reached. Set to 0 for unlimited retention. Message supports template variables like &#123;&#123;alarm_name&#125;&#125;.</small>
                    </div>
                    
                    <div id="jsonMessageSection" class="form-group message-input-section" style="display:none;">
                        <div class="message-header">
                            <label>📄 JSON File Configuration</label>
                            <div style="display: flex; gap: 8px; align-items: center;">
                                <select onchange="insertVariable('jsonMessage')" class="variable-dropdown">
                                    <option value="">📋 Insert Variable...</option>
                                    <option value="{{ "{{" }}app_info}}">{{ "{{" }}app_info}} - Application info (version, uptime)</option>
                                    <option value="{{ "{{" }}alarm_info}}">{{ "{{" }}alarm_info}} - Alarm info (name, desc, condition)</option>
                                    <option value="{{ "{{" }}sensor_info}}">{{ "{{" }}sensor_info}} - Sensor values that triggered alarm</option>
                                    <option value="{{ "{{" }}alarm_name}}">{{ "{{" }}alarm_name}} - Alarm name</option>
                                    <option value="{{ "{{" }}alarm_description}}">{{ "{{" }}alarm_description}} - Alarm description</option>
                                    <option value="{{ "{{" }}alarm_condition}}">{{ "{{" }}alarm_condition}} - Alarm condition</option>
                                    <option value="{{ "{{" }}station}}">{{ "{{" }}station}} - Station name</option>
                                    <option value="{{ "{{" }}timestamp}}">{{ "{{" }}timestamp}} - Current time</option>
                                    <option value="{{ "{{" }}temperature}}">{{ "{{" }}temperature}} - Temperature °C (current)</option>
                                    <option value="{{ "{{" }}temperature_f}}">{{ "{{" }}temperature_f}} - Temperature °F (current)</option>
                                    <option value="{{ "{{" }}humidity}}">{{ "{{" }}humidity}} - Humidity % (current)</option>
                                    <option value="{{ "{{" }}pressure}}">{{ "{{" }}pressure}} - Pressure mb (current)</option>
                                    <option value="{{ "{{" }}wind_speed}}">{{ "{{" }}wind_speed}} - Wind Speed m/s (current)</option>
                                    <option value="{{ "{{" }}wind_gust}}">{{ "{{" }}wind_gust}} - Wind Gust m/s (current)</option>
                                    <option value="{{ "{{" }}wind_direction}}">{{ "{{" }}wind_direction}} - Wind Direction° (current)</option>
                                    <option value="{{ "{{" }}lux}}">{{ "{{" }}lux}} - Light Lux (current)</option>
                                    <option value="{{ "{{" }}uv}}">{{ "{{" }}uv}} - UV Index (current)</option>
                                    <option value="{{ "{{" }}rain_rate}}">{{ "{{" }}rain_rate}} - Rain Rate mm (current)</option>
                                    <option value="{{ "{{" }}rain_daily}}">{{ "{{" }}rain_daily}} - Daily Rain mm (current)</option>
                                    <option value="{{ "{{" }}lightning_count}}">{{ "{{" }}lightning_count}} - Lightning Strikes (current)</option>
                                    <option value="{{ "{{" }}lightning_distance}}">{{ "{{" }}lightning_distance}} - Lightning Distance km (current)</option>
                                </select>
                                <button type="button" class="btn btn-secondary" onclick="showEmojiPicker('jsonMessage')" title="Insert Emoji">😀</button>
                            </div>
                        </div>
                        <label for="jsonPath" style="font-weight: 600;">File Path: <span style="color: red;">*</span></label>
                        <input type="text" id="jsonPath" placeholder="/tmp/tempest-alarms.json" />
                        <label for="jsonMaxDays" style="margin-top: 10px; font-weight: 600;">Max Days (0 = unlimited):</label>
                        <input type="number" id="jsonMaxDays" value="30" min="0" placeholder="30" />
                        <label for="jsonMessage" style="margin-top: 10px; font-weight: 600;">Message Template: <span style="color: red;">*</span></label>
                        <textarea id="jsonMessage" rows="3" placeholder="JSON message template..."></textarea>
                        <div style="display: flex; gap: 8px; margin-top: 8px;">
                            <button type="button" class="btn btn-info" onclick="validateJSONMessage()">✓ Validate JSON</button>
                            <button type="button" class="btn btn-secondary" onclick="showSampleJSON()">📄 Show Sample JSON</button>
                        </div>
                        <div id="jsonValidationResult" style="margin-top: 8px; padding: 8px; border-radius: 4px; display: none;"></div>
                        <small>JSON files will be rotated when max days is reached. Set to 0 for unlimited retention. Message supports template variables like &#123;&#123;alarm_name&#125;&#125;.</small>
                    </div>
                </div>
                
                <div class="form-group">
                    <label>🕐 Schedule (when alarm is active)</label>
                    <select id="scheduleType" onchange="toggleScheduleFields()">
                        <option value="">Always active (24/7)</option>
                        <option value="time">Time Range (daily)</option>
                        <option value="daily">Daily (same as time)</option>
                        <option value="weekly">Weekly (specific days)</option>
                        <option value="sun">Sunrise/Sunset based</option>
                    </select>
                    <small>Control when this alarm checks conditions</small>
                </div>
                
                <div id="timeScheduleSection" class="form-group" style="display:none; margin-left: 20px;">
                    <label>Start Time (HH:MM)</label>
                    <input type="time" id="scheduleStartTime" />
                    <label style="margin-top: 10px;">End Time (HH:MM)</label>
                    <input type="time" id="scheduleEndTime" />
                    <small>Active during this time range each day (supports overnight ranges like 22:00-06:00)</small>
                </div>
                
                <div id="weeklyScheduleSection" class="form-group" style="display:none; margin-left: 20px;">
                    <label>Active Days</label>
                    <div style="display: flex; gap: 10px; flex-wrap: wrap; margin-bottom: 10px;">
                        <label style="display: flex; align-items: center; gap: 5px;">
                            <input type="checkbox" class="schedule-day" value="0" /> Sunday
                        </label>
                        <label style="display: flex; align-items: center; gap: 5px;">
                            <input type="checkbox" class="schedule-day" value="1" /> Monday
                        </label>
                        <label style="display: flex; align-items: center; gap: 5px;">
                            <input type="checkbox" class="schedule-day" value="2" /> Tuesday
                        </label>
                        <label style="display: flex; align-items: center; gap: 5px;">
                            <input type="checkbox" class="schedule-day" value="3" /> Wednesday
                        </label>
                        <label style="display: flex; align-items: center; gap: 5px;">
                            <input type="checkbox" class="schedule-day" value="4" /> Thursday
                        </label>
                        <label style="display: flex; align-items: center; gap: 5px;">
                            <input type="checkbox" class="schedule-day" value="5" /> Friday
                        </label>
                        <label style="display: flex; align-items: center; gap: 5px;">
                            <input type="checkbox" class="schedule-day" value="6" /> Saturday
                        </label>
                    </div>
                    <label>
                        <input type="checkbox" id="weeklyTimeRange" onchange="toggleWeeklyTimeRange()" />
                        Also restrict to specific hours
                    </label>
                    <div id="weeklyTimeRangeFields" style="display:none; margin-top: 10px;">
                        <label>Start Time (HH:MM)</label>
                        <input type="time" id="weeklyStartTime" />
                        <label style="margin-top: 10px;">End Time (HH:MM)</label>
                        <input type="time" id="weeklyEndTime" />
                    </div>
                </div>
                
                <div id="sunScheduleSection" class="form-group" style="display:none; margin-left: 20px;">
                    <label>Start Event</label>
                    <select id="scheduleSunEvent">
                        <option value="sunrise">Sunrise</option>
                        <option value="sunset">Sunset</option>
                    </select>
                    <label style="margin-top: 10px;">Start Offset (minutes)</label>
                    <input type="number" id="scheduleSunOffset" value="0" placeholder="0" />
                    <small>Use negative for before (-30 = 30 min before), positive for after</small>
                    
                    <label style="margin-top: 10px;">
                        <input type="checkbox" id="sunHasEndEvent" onchange="toggleSunEndEvent()" />
                        Set end time (create a range)
                    </label>
                    <div id="sunEndEventFields" style="display:none; margin-top: 10px;">
                        <label>End Event</label>
                        <select id="scheduleSunEventEnd">
                            <option value="sunrise">Sunrise</option>
                            <option value="sunset">Sunset</option>
                        </select>
                        <label style="margin-top: 10px;">End Offset (minutes)</label>
                        <input type="number" id="scheduleSunOffsetEnd" value="0" placeholder="0" />
                    </div>
                    
                    <label style="margin-top: 10px;">
                        <input type="checkbox" id="scheduleUseStationLocation" />
                        Use weather station's location for sun calculations
                    </label>
                    <small>If unchecked, uses default location. If checked, sunrise/sunset times based on actual station location.</small>
                    
                    <div id="customLocationFields" style="margin-top: 10px;">
                        <label>
                            <input type="checkbox" id="sunHasCustomLocation" onchange="toggleCustomLocation()" />
                            Use custom location (overrides station location)
                        </label>
                        <div id="customLocationInputs" style="display:none; margin-top: 10px;">
                            <label>Latitude</label>
                            <input type="number" id="scheduleLatitude" step="0.0001" placeholder="34.0522" />
                            <label style="margin-top: 10px;">Longitude</label>
                            <input type="number" id="scheduleLongitude" step="0.0001" placeholder="-118.2437" />
                        </div>
                    </div>
                </div>
                
                <div class="form-group">
                    <label>Tags</label>
                    <div class="tag-selector-container">
                        <div class="selected-tags" id="selectedTags"></div>
                        <div class="tag-dropdown-wrapper">
                            <input type="text" 
                                   id="tagSearchInput" 
                                   class="tag-search-input" 
                                   placeholder="Search or create tags..." 
                                   autocomplete="off" />
                            <div id="tagDropdown" class="tag-dropdown"></div>
                        </div>
                    </div>
                </div>
                
                <div class="form-group">
                    <label>Cooldown (seconds)</label>
                    <input type="number" id="alarmCooldown" value="1800" min="0" />
                    <small>Minimum time between consecutive alarm triggers</small>
                </div>
                
                <div class="form-group">
                    <label>
                        <input type="checkbox" id="alarmEnabled" checked />
                        Enabled
                    </label>
                </div>
                
                <div class="modal-actions">
                    <button type="button" class="btn btn-secondary" onclick="closeModal()">Cancel</button>
                    <button type="button" class="btn btn-danger" onclick="deleteAlarm()" id="deleteBtn" style="display:none;">Delete</button>
                    <button type="submit" class="btn btn-primary">Save</button>
                </div>
            </form>
        </div>
    </div>
    
    <div id="notification" class="notification"></div>
    
    <div id="emojiModal" class="modal">
        <div class="modal-content">
            <div class="modal-header">😀 Choose Emoji</div>
            <div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(40px, 1fr)); gap: 8px; max-height: 300px; overflow-y: auto; padding: 20px;">
                <button onclick="insertEmoji('consoleMessage', '⚡')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">⚡</button>
                <button onclick="insertEmoji('consoleMessage', '🌡️')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">🌡️</button>
                <button onclick="insertEmoji('consoleMessage', '💧')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">💧</button>
                <button onclick="insertEmoji('consoleMessage', '🌬️')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">🌬️</button>
                <button onclick="insertEmoji('consoleMessage', '🌪️')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">🌪️</button>
                <button onclick="insertEmoji('consoleMessage', '🌧️')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">🌧️</button>
                <button onclick="insertEmoji('consoleMessage', '⛈️')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">⛈️</button>
                <button onclick="insertEmoji('consoleMessage', '🌩️')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">🌩️</button>
                <button onclick="insertEmoji('consoleMessage', '☀️')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">☀️</button>
                <button onclick="insertEmoji('consoleMessage', '🌤️')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">🌤️</button>
                <button onclick="insertEmoji('consoleMessage', '⛅')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">⛅</button>
                <button onclick="insertEmoji('consoleMessage', '☁️')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">☁️</button>
                <button onclick="insertEmoji('consoleMessage', '🌫️')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">🌫️</button>
                <button onclick="insertEmoji('consoleMessage', '🌈')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">🌈</button>
                <button onclick="insertEmoji('consoleMessage', '❄️')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">❄️</button>
                <button onclick="insertEmoji('consoleMessage', '🌊')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">🌊</button>
                <button onclick="insertEmoji('consoleMessage', '🌍')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">🌍</button>
                <button onclick="insertEmoji('consoleMessage', '📊')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">📊</button>
                <button onclick="insertEmoji('consoleMessage', '📈')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">📈</button>
                <button onclick="insertEmoji('consoleMessage', '📉')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">📉</button>
                <button onclick="insertEmoji('consoleMessage', '🔥')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">🔥</button>
                <button onclick="insertEmoji('consoleMessage', '🔔')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">🔔</button>
                <button onclick="insertEmoji('consoleMessage', '⚠️')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">⚠️</button>
                <button onclick="insertEmoji('consoleMessage', '🚨')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">🚨</button>
                <button onclick="insertEmoji('consoleMessage', '✅')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">✅</button>
                <button onclick="insertEmoji('consoleMessage', '❌')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">❌</button>
                <button onclick="insertEmoji('consoleMessage', '⏰')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">⏰</button>
                <button onclick="insertEmoji('consoleMessage', '📅')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">📅</button>
                <button onclick="insertEmoji('consoleMessage', '🏠')" style="font-size: 24px; border: none; background: none; cursor: pointer; padding: 8px;">🏠</button>
            </div>
            <div class="modal-actions">
                <button type="button" class="btn btn-secondary" onclick="closeEmojiModal()">Close</button>
            </div>
        </div>
    </div>

    <div id="editContactsModal" class="modal">
        <div class="modal-content wide">
            <div class="modal-header">👥 Edit Contact List</div>
            <div class="contacts-editor">
                <div class="contacts-list" id="contactsList"></div>
                <button class="btn btn-primary" onclick="addNewContact()" style="margin: 10px 0;">+ Add Contact</button>
            </div>
            <div class="modal-actions">
                <button type="button" class="btn btn-secondary" onclick="closeEditContactsModal()">Cancel</button>
                <button type="button" class="btn btn-success" onclick="saveContacts('json')">💾 Save as JSON File</button>
                <button type="button" class="btn btn-info" onclick="saveContacts('env')">📝 Update {{.EnvFile}} File</button>
            </div>
        </div>
    </div>

    <div id="editTagsModal" class="modal">
        <div class="modal-content">
            <div class="modal-header">🏷️ Edit Tag List</div>
            <div class="tags-editor">
                <div class="tags-list" id="tagsList"></div>
                <button class="btn btn-primary" onclick="addNewTag()" style="margin: 10px 0;">+ Add Tag</button>
            </div>
            <div class="modal-actions">
                <button type="button" class="btn btn-secondary" onclick="closeEditTagsModal()">Cancel</button>
                <button type="button" class="btn btn-success" onclick="saveTags('json')">💾 Save as JSON File</button>
                <button type="button" class="btn btn-info" onclick="saveTags('env')">📝 Update {{.EnvFile}} File</button>
            </div>
        </div>
    </div>
    
    <script src="/alarm-editor/static/script.js"></script>
    
    <div class="footer">
        <p>Last updated: <span id="last-update">--</span></p>
        <p>Tempest HomeKit Service v{{.Version}}</p>
        <div class="theme-selector">
            <label for="theme-select">🎨 Theme:</label>
            <select id="theme-select">
                <option value="default">Default (Purple)</option>
                <option value="ocean">Ocean Blue</option>
                <option value="sunset">Sunset Orange</option>
                <option value="forest">Forest Green</option>
                <option value="midnight">Midnight Dark</option>
                <option value="arctic">Arctic Light</option>
                <option value="autumn">Autumn Earth</option>
            </select>
        </div>
    </div>
</body>
</html>
`
