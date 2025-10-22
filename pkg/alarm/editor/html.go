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
                    </div>
                    <small>Select at least one delivery method. Each method will show its configuration below with defaults pre-populated.</small>
                </div>
                
                <div id="messageSections">
                    <div id="consoleMessageSection" class="form-group message-input-section" style="display:block;">
                        <div class="message-header">
                            <label>📟 Console Message</label>
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
                        </div>
                        <textarea id="consoleMessage" rows="5" placeholder="Console-specific message..."></textarea>
                    </div>
                    
                    <div id="syslogMessageSection" class="form-group message-input-section" style="display:none;">
                        <div class="message-header">
                            <label>📋 Syslog Message</label>
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
                        </div>
                        <textarea id="syslogMessage" rows="5" placeholder="Syslog-specific message..."></textarea>
                    </div>
                    
                    <div id="oslogMessageSection" class="form-group message-input-section" style="display:none;">
                        <div class="message-header">
                            <label>🍎 OSLog Message (macOS Unified Logging)</label>
                            <select onchange="insertVariable('oslogMessage')" class="variable-dropdown">
                                <option value="">📋 Insert Variable...</option>
                                <option value="{{ "{{" }}app_info}}">{{ "{{" }}app_info}} - Application info (version, uptime)</option>
                                <option value="{{ "{{" }}alarm_info}}">{{ "{{" }}alarm_info}} - Alarm info (name, desc, condition)</option>
                                <option value="{{ "{{" }}sensor_info}}">{{ "{{" }}sensor_info}} - Sensor values that triggered alarm</option>
                                <option value="{{ "{{" }}alarm_name}}">{{ "{{" }}alarm_name}} - Alarm name</option>
                                <option value="{{ "{{" }}alarm_description}}">{{ "{{" }}alarm_description}} - Alarm description</option>
                            </select>
                        </div>
                        <textarea id="oslogMessage" rows="5" placeholder="OSLog-specific message (visible in Console.app and log command)..."></textarea>
                    </div>
                    
                    <div id="eventlogMessageSection" class="form-group message-input-section" style="display:none;">
                        <div class="message-header">
                            <label>📊 Event Log Message</label>
                            <select onchange="insertVariable('eventlogMessage')" class="variable-dropdown">
                                <option value="">📋 Insert Variable...</option>
                                <option value="{{ "{{" }}app_info}}">{{ "{{" }}app_info}} - Application info (version, uptime)</option>
                                <option value="{{ "{{" }}alarm_info}}">{{ "{{" }}alarm_info}} - Alarm info (name, desc, condition)</option>
                                <option value="{{ "{{" }}sensor_info}}">{{ "{{" }}sensor_info}} - Sensor values that triggered alarm</option>
                                <option value="{{ "{{" }}alarm_name}}">{{ "{{" }}alarm_name}} - Alarm name</option>
                                <option value="{{ "{{" }}alarm_description}}">{{ "{{" }}alarm_description}} - Alarm description</option>
                            </select>
                        </div>
                        <textarea id="eventlogMessage" rows="5" placeholder="Event log-specific message..."></textarea>
                    </div>
                    
                    <div id="emailMessageSection" class="form-group message-input-section" style="display:none;">
                        <div class="message-header">
                            <label>✉️ Email Configuration</label>
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
                        </div>
                        <label for="emailTo" style="font-weight: 600;">To: <span style="color: red;">*</span></label>
                        <input type="text" id="emailTo" placeholder="recipient@example.com (comma-separated for multiple)" />
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
                            <select onchange="insertVariable('smsMessage')" class="variable-dropdown">
                                <option value="">📋 Insert Variable...</option>
                                <option value="{{ "{{" }}app_info}}">{{ "{{" }}app_info}} - Application info (version, uptime)</option>
                                <option value="{{ "{{" }}alarm_info}}">{{ "{{" }}alarm_info}} - Alarm info (name, desc, condition)</option>
                                <option value="{{ "{{" }}sensor_info}}">{{ "{{" }}sensor_info}} - Sensor values that triggered alarm</option>
                                <option value="{{ "{{" }}alarm_name}}">{{ "{{" }}alarm_name}} - Alarm name</option>
                                <option value="{{ "{{" }}station}}">{{ "{{" }}station}} - Station name</option>
                                <option value="{{ "{{" }}timestamp}}">{{ "{{" }}timestamp}} - Current time</option>
                            </select>
                        </div>
                        <label for="smsTo" style="font-weight: 600;">To: <span style="color: red;">*</span></label>
                        <input type="text" id="smsTo" placeholder="Phone number(s) (comma-separated)" />
                        <label for="smsMessage" style="margin-top: 10px; font-weight: 600;">Message:</label>
                        <textarea id="smsMessage" rows="3" placeholder="SMS message (keep it short)..."></textarea>
                    </div>
                    
                    <div id="webhookMessageSection" class="form-group message-input-section" style="display:none;">
                        <div class="message-header">
                            <label>🌐 Webhook Configuration</label>
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
