package editor

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Alarm Configuration Editor</title>
    <link rel="stylesheet" href="/alarm-editor/static/styles.css">
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>⚡ Tempest Alarm Editor</h1>
            <p>Create and manage weather alarms with real-time monitoring</p>
            <div class="config-path-display">
                <span class="label">📁 Watching:</span>
                <span class="path">{{"{{"}}{{.ConfigPath}}</span>
                <span class="label" style="margin-left: 20px;">🕒 Last read:</span>
                <span class="path">{{"{{"}}{{.LastLoad}}</span>
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
                    </div>
                    <small>Select at least one delivery method</small>
                </div>
                
                <div class="form-group message-section">
                    <label>
                        <input type="checkbox" id="useDefaultMessage" checked onchange="toggleCustomMessages()" />
                        Use Default Message for All Selected Methods
                    </label>
                    <div id="defaultMessageSection" class="message-input-section">
                        <div class="message-header">
                            <label>Default Message Template</label>
                            <select id="variableDropdown" onchange="insertVariable('defaultMessage')" class="variable-dropdown">
                                <option value="">📋 Insert Variable...</option>
                                <option value="{{"{{"}}alarm_name}}">{{"{{"}}alarm_name}} - Alarm name</option>
                                <option value="{{"{{"}}alarm_description}}">{{"{{"}}alarm_description}} - Alarm description</option>
                                <option value="{{"{{"}}station}}">{{"{{"}}station}} - Station name</option>
                                <option value="{{"{{"}}timestamp}}">{{"{{"}}timestamp}} - Current time</option>
                                <option value="{{"{{"}}temperature}}">{{"{{"}}temperature}} - Temperature °C (current)</option>
                                <option value="{{"{{"}}temperature_f}}">{{"{{"}}temperature_f}} - Temperature °F (current)</option>
                                <option value="{{"{{"}}temperature_c}}">{{"{{"}}temperature_c}} - Temperature °C (current)</option>
                                <option value="{{"{{"}}humidity}}">{{"{{"}}humidity}} - Humidity % (current)</option>
                                <option value="{{"{{"}}pressure}}">{{"{{"}}pressure}} - Pressure mb (current)</option>
                                <option value="{{"{{"}}wind_speed}}">{{"{{"}}wind_speed}} - Wind Speed m/s (current)</option>
                                <option value="{{"{{"}}wind_gust}}">{{"{{"}}wind_gust}} - Wind Gust m/s (current)</option>
                                <option value="{{"{{"}}wind_direction}}">{{"{{"}}wind_direction}} - Wind Direction° (current)</option>
                                <option value="{{"{{"}}lux}}">{{"{{"}}lux}} - Light Lux (current)</option>
                                <option value="{{"{{"}}uv}}">{{"{{"}}uv}} - UV Index (current)</option>
                                <option value="{{"{{"}}rain_rate}}">{{"{{"}}rain_rate}} - Rain Rate mm (current)</option>
                                <option value="{{"{{"}}rain_daily}}">{{"{{"}}rain_daily}} - Daily Rain mm (current)</option>
                                <option value="{{"{{"}}lightning_count}}">{{"{{"}}lightning_count}} - Lightning Strikes (current)</option>
                                <option value="{{"{{"}}lightning_distance}}">{{"{{"}}lightning_distance}} - Lightning Distance km (current)</option>
                                <option value="{{"{{"}}last_temperature}}">{{"{{"}}last_temperature}} - Temperature °C (previous)</option>
                                <option value="{{"{{"}}last_humidity}}">{{"{{"}}last_humidity}} - Humidity % (previous)</option>
                                <option value="{{"{{"}}last_pressure}}">{{"{{"}}last_pressure}} - Pressure mb (previous)</option>
                                <option value="{{"{{"}}last_wind_speed}}">{{"{{"}}last_wind_speed}} - Wind Speed m/s (previous)</option>
                                <option value="{{"{{"}}last_wind_gust}}">{{"{{"}}last_wind_gust}} - Wind Gust m/s (previous)</option>
                                <option value="{{"{{"}}last_wind_direction}}">{{"{{"}}last_wind_direction}} - Wind Direction° (previous)</option>
                                <option value="{{"{{"}}last_lux}}">{{"{{"}}last_lux}} - Light Lux (previous)</option>
                                <option value="{{"{{"}}last_uv}}">{{"{{"}}last_uv}} - UV Index (previous)</option>
                                <option value="{{"{{"}}last_rain_rate}}">{{"{{"}}last_rain_rate}} - Rain Rate mm (previous)</option>
                                <option value="{{"{{"}}last_rain_daily}}">{{"{{"}}last_rain_daily}} - Daily Rain mm (previous)</option>
                                <option value="{{"{{"}}last_lightning_count}}">{{"{{"}}last_lightning_count}} - Lightning Strikes (previous)</option>
                                <option value="{{"{{"}}last_lightning_distance}}">{{"{{"}}last_lightning_distance}} - Lightning Distance km (previous)</option>
                            </select>
                        </div>
                        <textarea id="defaultMessage" rows="5" placeholder="Enter default message template for all channels...">🚨 ALARM: {{"{{"}}alarm_name}}
{{"{{"}}alarm_description}}

Station: {{"{{"}}station}}
Time: {{"{{"}}timestamp}}</textarea>
                    </div>
                </div>
                
                <div id="customMessageSections" style="display:none;">
                    <div id="consoleMessageSection" class="message-input-section" style="display:none;">
                        <div class="message-header">
                            <label>📟 Console Message</label>
                            <select onchange="insertVariable('consoleMessage')" class="variable-dropdown">
                                <option value="">📋 Insert Variable...</option>
                                <option value="{{"{{"}}alarm_name}}">{{"{{"}}alarm_name}} - Alarm name</option>
                                <option value="{{"{{"}}alarm_description}}">{{"{{"}}alarm_description}} - Alarm description</option>
                                <option value="{{"{{"}}station}}">{{"{{"}}station}} - Station name</option>
                                <option value="{{"{{"}}timestamp}}">{{"{{"}}timestamp}} - Current time</option>
                                <option value="{{"{{"}}temperature}}">{{"{{"}}temperature}} - Temperature °C (current)</option>
                                <option value="{{"{{"}}temperature_f}}">{{"{{"}}temperature_f}} - Temperature °F (current)</option>
                                <option value="{{"{{"}}humidity}}">{{"{{"}}humidity}} - Humidity % (current)</option>
                                <option value="{{"{{"}}pressure}}">{{"{{"}}pressure}} - Pressure mb (current)</option>
                                <option value="{{"{{"}}wind_speed}}">{{"{{"}}wind_speed}} - Wind Speed m/s (current)</option>
                                <option value="{{"{{"}}wind_gust}}">{{"{{"}}wind_gust}} - Wind Gust m/s (current)</option>
                                <option value="{{"{{"}}wind_direction}}">{{"{{"}}wind_direction}} - Wind Direction° (current)</option>
                                <option value="{{"{{"}}lux}}">{{"{{"}}lux}} - Light Lux (current)</option>
                                <option value="{{"{{"}}uv}}">{{"{{"}}uv}} - UV Index (current)</option>
                                <option value="{{"{{"}}rain_rate}}">{{"{{"}}rain_rate}} - Rain Rate mm (current)</option>
                                <option value="{{"{{"}}lightning_count}}">{{"{{"}}lightning_count}} - Lightning Strikes (current)</option>
                                <option value="{{"{{"}}lightning_distance}}">{{"{{"}}lightning_distance}} - Lightning Distance km (current)</option>
                                <option value="{{"{{"}}last_temperature}}">{{"{{"}}last_temperature}} - Temperature °C (previous)</option>
                                <option value="{{"{{"}}last_humidity}}">{{"{{"}}last_humidity}} - Humidity % (previous)</option>
                                <option value="{{"{{"}}last_pressure}}">{{"{{"}}last_pressure}} - Pressure mb (previous)</option>
                                <option value="{{"{{"}}last_wind_speed}}">{{"{{"}}last_wind_speed}} - Wind Speed m/s (previous)</option>
                                <option value="{{"{{"}}last_wind_gust}}">{{"{{"}}last_wind_gust}} - Wind Gust m/s (previous)</option>
                                <option value="{{"{{"}}last_wind_direction}}">{{"{{"}}last_wind_direction}} - Wind Direction° (previous)</option>
                                <option value="{{"{{"}}last_lux}}">{{"{{"}}last_lux}} - Light Lux (previous)</option>
                                <option value="{{"{{"}}last_uv}}">{{"{{"}}last_uv}} - UV Index (previous)</option>
                                <option value="{{"{{"}}last_rain_rate}}">{{"{{"}}last_rain_rate}} - Rain Rate mm (previous)</option>
                                <option value="{{"{{"}}last_lightning_count}}">{{"{{"}}last_lightning_count}} - Lightning Strikes (previous)</option>
                                <option value="{{"{{"}}last_lightning_distance}}">{{"{{"}}last_lightning_distance}} - Lightning Distance km (previous)</option>
                            </select>
                        </div>
                        <textarea id="consoleMessage" rows="4" placeholder="Console-specific message..."></textarea>
                    </div>
                    
                    <div id="syslogMessageSection" class="message-input-section" style="display:none;">
                        <div class="message-header">
                            <label>📋 Syslog Message</label>
                            <select onchange="insertVariable('syslogMessage')" class="variable-dropdown">
                                <option value="">📋 Insert Variable...</option>
                                <option value="{{"{{"}}alarm_name}}">{{"{{"}}alarm_name}} - Alarm name</option>
                                <option value="{{"{{"}}alarm_description}}">{{"{{"}}alarm_description}} - Alarm description</option>
                                <option value="{{"{{"}}station}}">{{"{{"}}station}} - Station name</option>
                                <option value="{{"{{"}}timestamp}}">{{"{{"}}timestamp}} - Current time</option>
                                <option value="{{"{{"}}temperature}}">{{"{{"}}temperature}} - Temperature °C (current)</option>
                                <option value="{{"{{"}}temperature_f}}">{{"{{"}}temperature_f}} - Temperature °F (current)</option>
                                <option value="{{"{{"}}humidity}}">{{"{{"}}humidity}} - Humidity % (current)</option>
                                <option value="{{"{{"}}pressure}}">{{"{{"}}pressure}} - Pressure mb (current)</option>
                                <option value="{{"{{"}}wind_speed}}">{{"{{"}}wind_speed}} - Wind Speed m/s (current)</option>
                                <option value="{{"{{"}}wind_gust}}">{{"{{"}}wind_gust}} - Wind Gust m/s (current)</option>
                                <option value="{{"{{"}}wind_direction}}">{{"{{"}}wind_direction}} - Wind Direction° (current)</option>
                                <option value="{{"{{"}}lux}}">{{"{{"}}lux}} - Light Lux (current)</option>
                                <option value="{{"{{"}}uv}}">{{"{{"}}uv}} - UV Index (current)</option>
                                <option value="{{"{{"}}rain_rate}}">{{"{{"}}rain_rate}} - Rain Rate mm (current)</option>
                                <option value="{{"{{"}}lightning_count}}">{{"{{"}}lightning_count}} - Lightning Strikes (current)</option>
                                <option value="{{"{{"}}lightning_distance}}">{{"{{"}}lightning_distance}} - Lightning Distance km (current)</option>
                                <option value="{{"{{"}}last_temperature}}">{{"{{"}}last_temperature}} - Temperature °C (previous)</option>
                                <option value="{{"{{"}}last_humidity}}">{{"{{"}}last_humidity}} - Humidity % (previous)</option>
                                <option value="{{"{{"}}last_pressure}}">{{"{{"}}last_pressure}} - Pressure mb (previous)</option>
                                <option value="{{"{{"}}last_wind_speed}}">{{"{{"}}last_wind_speed}} - Wind Speed m/s (previous)</option>
                                <option value="{{"{{"}}last_wind_gust}}">{{"{{"}}last_wind_gust}} - Wind Gust m/s (previous)</option>
                                <option value="{{"{{"}}last_wind_direction}}">{{"{{"}}last_wind_direction}} - Wind Direction° (previous)</option>
                                <option value="{{"{{"}}last_lux}}">{{"{{"}}last_lux}} - Light Lux (previous)</option>
                                <option value="{{"{{"}}last_uv}}">{{"{{"}}last_uv}} - UV Index (previous)</option>
                                <option value="{{"{{"}}last_rain_rate}}">{{"{{"}}last_rain_rate}} - Rain Rate mm (previous)</option>
                                <option value="{{"{{"}}last_lightning_count}}">{{"{{"}}last_lightning_count}} - Lightning Strikes (previous)</option>
                                <option value="{{"{{"}}last_lightning_distance}}">{{"{{"}}last_lightning_distance}} - Lightning Distance km (previous)</option>
                            </select>
                        </div>
                        <textarea id="syslogMessage" rows="4" placeholder="Syslog-specific message..."></textarea>
                    </div>
                    
                    <div id="oslogMessageSection" class="message-input-section" style="display:none;">
                        <div class="message-header">
                            <label>🍎 OSLog Message (macOS Unified Logging)</label>
                            <select onchange="insertVariable('oslogMessage')" class="variable-dropdown">
                                <option value="">📋 Insert Variable...</option>
                                <option value="{{"{{"}}alarm_name}}">{{"{{"}}alarm_name}} - Alarm name</option>
                                <option value="{{"{{"}}alarm_description}}">{{"{{"}}alarm_description}} - Alarm description</option>
                                <option value="{{"{{"}}station}}">{{"{{"}}station}} - Station name</option>
                                <option value="{{"{{"}}timestamp}}">{{"{{"}}timestamp}} - Current time</option>
                                <option value="{{"{{"}}temperature}}">{{"{{"}}temperature}} - Temperature °C (current)</option>
                                <option value="{{"{{"}}temperature_f}}">{{"{{"}}temperature_f}} - Temperature °F (current)</option>
                                <option value="{{"{{"}}humidity}}">{{"{{"}}humidity}} - Humidity % (current)</option>
                                <option value="{{"{{"}}pressure}}">{{"{{"}}pressure}} - Pressure mb (current)</option>
                                <option value="{{"{{"}}wind_speed}}">{{"{{"}}wind_speed}} - Wind Speed m/s (current)</option>
                                <option value="{{"{{"}}wind_gust}}">{{"{{"}}wind_gust}} - Wind Gust m/s (current)</option>
                                <option value="{{"{{"}}wind_direction}}">{{"{{"}}wind_direction}} - Wind Direction° (current)</option>
                                <option value="{{"{{"}}lux}}">{{"{{"}}lux}} - Light Lux (current)</option>
                                <option value="{{"{{"}}uv}}">{{"{{"}}uv}} - UV Index (current)</option>
                                <option value="{{"{{"}}rain_rate}}">{{"{{"}}rain_rate}} - Rain Rate mm (current)</option>
                                <option value="{{"{{"}}lightning_count}}">{{"{{"}}lightning_count}} - Lightning Strikes (current)</option>
                                <option value="{{"{{"}}lightning_distance}}">{{"{{"}}lightning_distance}} - Lightning Distance km (current)</option>
                                <option value="{{"{{"}}last_temperature}}">{{"{{"}}last_temperature}} - Temperature °C (previous)</option>
                                <option value="{{"{{"}}last_humidity}}">{{"{{"}}last_humidity}} - Humidity % (previous)</option>
                                <option value="{{"{{"}}last_pressure}}">{{"{{"}}last_pressure}} - Pressure mb (previous)</option>
                                <option value="{{"{{"}}last_wind_speed}}">{{"{{"}}last_wind_speed}} - Wind Speed m/s (previous)</option>
                                <option value="{{"{{"}}last_wind_gust}}">{{"{{"}}last_wind_gust}} - Wind Gust m/s (previous)</option>
                                <option value="{{"{{"}}last_wind_direction}}">{{"{{"}}last_wind_direction}} - Wind Direction° (previous)</option>
                                <option value="{{"{{"}}last_lux}}">{{"{{"}}last_lux}} - Light Lux (previous)</option>
                                <option value="{{"{{"}}last_uv}}">{{"{{"}}last_uv}} - UV Index (previous)</option>
                                <option value="{{"{{"}}last_rain_rate}}">{{"{{"}}last_rain_rate}} - Rain Rate mm (previous)</option>
                                <option value="{{"{{"}}last_lightning_count}}">{{"{{"}}last_lightning_count}} - Lightning Strikes (previous)</option>
                                <option value="{{"{{"}}last_lightning_distance}}">{{"{{"}}last_lightning_distance}} - Lightning Distance km (previous)</option>
                            </select>
                        </div>
                        <textarea id="oslogMessage" rows="4" placeholder="OSLog-specific message (visible in Console.app and log command)..."></textarea>
                    </div>
                    
                    <div id="eventlogMessageSection" class="message-input-section" style="display:none;">
                        <div class="message-header">
                            <label>📊 Event Log Message</label>
                            <select onchange="insertVariable('eventlogMessage')" class="variable-dropdown">
                                <option value="">📋 Insert Variable...</option>
                                <option value="{{"{{"}}alarm_name}}">{{"{{"}}alarm_name}} - Alarm name</option>
                                <option value="{{"{{"}}alarm_description}}">{{"{{"}}alarm_description}} - Alarm description</option>
                                <option value="{{"{{"}}station}}">{{"{{"}}station}} - Station name</option>
                                <option value="{{"{{"}}timestamp}}">{{"{{"}}timestamp}} - Current time</option>
                                <option value="{{"{{"}}temperature}}">{{"{{"}}temperature}} - Temperature °C (current)</option>
                                <option value="{{"{{"}}temperature_f}}">{{"{{"}}temperature_f}} - Temperature °F (current)</option>
                                <option value="{{"{{"}}humidity}}">{{"{{"}}humidity}} - Humidity % (current)</option>
                                <option value="{{"{{"}}pressure}}">{{"{{"}}pressure}} - Pressure mb (current)</option>
                                <option value="{{"{{"}}wind_speed}}">{{"{{"}}wind_speed}} - Wind Speed m/s (current)</option>
                                <option value="{{"{{"}}wind_gust}}">{{"{{"}}wind_gust}} - Wind Gust m/s (current)</option>
                                <option value="{{"{{"}}wind_direction}}">{{"{{"}}wind_direction}} - Wind Direction° (current)</option>
                                <option value="{{"{{"}}lux}}">{{"{{"}}lux}} - Light Lux (current)</option>
                                <option value="{{"{{"}}uv}}">{{"{{"}}uv}} - UV Index (current)</option>
                                <option value="{{"{{"}}rain_rate}}">{{"{{"}}rain_rate}} - Rain Rate mm (current)</option>
                                <option value="{{"{{"}}lightning_count}}">{{"{{"}}lightning_count}} - Lightning Strikes (current)</option>
                                <option value="{{"{{"}}lightning_distance}}">{{"{{"}}lightning_distance}} - Lightning Distance km (current)</option>
                                <option value="{{"{{"}}last_temperature}}">{{"{{"}}last_temperature}} - Temperature °C (previous)</option>
                                <option value="{{"{{"}}last_humidity}}">{{"{{"}}last_humidity}} - Humidity % (previous)</option>
                                <option value="{{"{{"}}last_pressure}}">{{"{{"}}last_pressure}} - Pressure mb (previous)</option>
                                <option value="{{"{{"}}last_wind_speed}}">{{"{{"}}last_wind_speed}} - Wind Speed m/s (previous)</option>
                                <option value="{{"{{"}}last_wind_gust}}">{{"{{"}}last_wind_gust}} - Wind Gust m/s (previous)</option>
                                <option value="{{"{{"}}last_wind_direction}}">{{"{{"}}last_wind_direction}} - Wind Direction° (previous)</option>
                                <option value="{{"{{"}}last_lux}}">{{"{{"}}last_lux}} - Light Lux (previous)</option>
                                <option value="{{"{{"}}last_uv}}">{{"{{"}}last_uv}} - UV Index (previous)</option>
                                <option value="{{"{{"}}last_rain_rate}}">{{"{{"}}last_rain_rate}} - Rain Rate mm (previous)</option>
                                <option value="{{"{{"}}last_lightning_count}}">{{"{{"}}last_lightning_count}} - Lightning Strikes (previous)</option>
                                <option value="{{"{{"}}last_lightning_distance}}">{{"{{"}}last_lightning_distance}} - Lightning Distance km (previous)</option>
                            </select>
                        </div>
                        <textarea id="eventlogMessage" rows="4" placeholder="Event log-specific message..."></textarea>
                    </div>
                    
                    <div id="emailMessageSection" class="message-input-section" style="display:none;">
                        <div class="message-header">
                            <label>✉️ Email Configuration</label>
                            <select onchange="insertVariable('emailSubject', 'emailBody')" class="variable-dropdown">
                                <option value="">📋 Insert Variable...</option>
                                <option value="{{"{{"}}alarm_name}}">{{"{{"}}alarm_name}} - Alarm name</option>
                                <option value="{{"{{"}}alarm_description}}">{{"{{"}}alarm_description}} - Alarm description</option>
                                <option value="{{"{{"}}station}}">{{"{{"}}station}} - Station name</option>
                                <option value="{{"{{"}}timestamp}}">{{"{{"}}timestamp}} - Current time</option>
                                <option value="{{"{{"}}temperature}}">{{"{{"}}temperature}} - Temperature °C (current)</option>
                                <option value="{{"{{"}}temperature_f}}">{{"{{"}}temperature_f}} - Temperature °F (current)</option>
                                <option value="{{"{{"}}humidity}}">{{"{{"}}humidity}} - Humidity % (current)</option>
                                <option value="{{"{{"}}pressure}}">{{"{{"}}pressure}} - Pressure mb (current)</option>
                                <option value="{{"{{"}}wind_speed}}">{{"{{"}}wind_speed}} - Wind Speed m/s (current)</option>
                                <option value="{{"{{"}}wind_gust}}">{{"{{"}}wind_gust}} - Wind Gust m/s (current)</option>
                                <option value="{{"{{"}}wind_direction}}">{{"{{"}}wind_direction}} - Wind Direction° (current)</option>
                                <option value="{{"{{"}}lux}}">{{"{{"}}lux}} - Light Lux (current)</option>
                                <option value="{{"{{"}}uv}}">{{"{{"}}uv}} - UV Index (current)</option>
                                <option value="{{"{{"}}rain_rate}}">{{"{{"}}rain_rate}} - Rain Rate mm (current)</option>
                                <option value="{{"{{"}}lightning_count}}">{{"{{"}}lightning_count}} - Lightning Strikes (current)</option>
                                <option value="{{"{{"}}lightning_distance}}">{{"{{"}}lightning_distance}} - Lightning Distance km (current)</option>
                                <option value="{{"{{"}}last_temperature}}">{{"{{"}}last_temperature}} - Temperature °C (previous)</option>
                                <option value="{{"{{"}}last_humidity}}">{{"{{"}}last_humidity}} - Humidity % (previous)</option>
                                <option value="{{"{{"}}last_pressure}}">{{"{{"}}last_pressure}} - Pressure mb (previous)</option>
                                <option value="{{"{{"}}last_wind_speed}}">{{"{{"}}last_wind_speed}} - Wind Speed m/s (previous)</option>
                                <option value="{{"{{"}}last_wind_gust}}">{{"{{"}}last_wind_gust}} - Wind Gust m/s (previous)</option>
                                <option value="{{"{{"}}last_wind_direction}}">{{"{{"}}last_wind_direction}} - Wind Direction° (previous)</option>
                                <option value="{{"{{"}}last_lux}}">{{"{{"}}last_lux}} - Light Lux (previous)</option>
                                <option value="{{"{{"}}last_uv}}">{{"{{"}}last_uv}} - UV Index (previous)</option>
                                <option value="{{"{{"}}last_rain_rate}}">{{"{{"}}last_rain_rate}} - Rain Rate mm (previous)</option>
                                <option value="{{"{{"}}last_lightning_count}}">{{"{{"}}last_lightning_count}} - Lightning Strikes (previous)</option>
                                <option value="{{"{{"}}last_lightning_distance}}">{{"{{"}}last_lightning_distance}} - Lightning Distance km (previous)</option>
                            </select>
                        </div>
                        <input type="text" id="emailTo" placeholder="Recipient email (comma-separated)" />
                        <input type="text" id="emailSubject" placeholder="Subject" />
                        <textarea id="emailBody" rows="4" placeholder="Email body..."></textarea>
                    </div>
                    
                    <div id="smsMessageSection" class="message-input-section" style="display:none;">
                        <div class="message-header">
                            <label>📱 SMS Configuration</label>
                            <select onchange="insertVariable('smsMessage')" class="variable-dropdown">
                                <option value="">📋 Insert Variable...</option>
                                <option value="{{"{{"}}alarm_name}}">{{"{{"}}alarm_name}} - Alarm name</option>
                                <option value="{{"{{"}}alarm_description}}">{{"{{"}}alarm_description}} - Alarm description</option>
                                <option value="{{"{{"}}station}}">{{"{{"}}station}} - Station name</option>
                                <option value="{{"{{"}}timestamp}}">{{"{{"}}timestamp}} - Current time</option>
                                <option value="{{"{{"}}temperature}}">{{"{{"}}temperature}} - Temperature °C (current)</option>
                                <option value="{{"{{"}}temperature_f}}">{{"{{"}}temperature_f}} - Temperature °F (current)</option>
                                <option value="{{"{{"}}humidity}}">{{"{{"}}humidity}} - Humidity % (current)</option>
                                <option value="{{"{{"}}pressure}}">{{"{{"}}pressure}} - Pressure mb (current)</option>
                                <option value="{{"{{"}}wind_speed}}">{{"{{"}}wind_speed}} - Wind Speed m/s (current)</option>
                                <option value="{{"{{"}}wind_gust}}">{{"{{"}}wind_gust}} - Wind Gust m/s (current)</option>
                                <option value="{{"{{"}}wind_direction}}">{{"{{"}}wind_direction}} - Wind Direction° (current)</option>
                                <option value="{{"{{"}}lux}}">{{"{{"}}lux}} - Light Lux (current)</option>
                                <option value="{{"{{"}}uv}}">{{"{{"}}uv}} - UV Index (current)</option>
                                <option value="{{"{{"}}rain_rate}}">{{"{{"}}rain_rate}} - Rain Rate mm (current)</option>
                                <option value="{{"{{"}}lightning_count}}">{{"{{"}}lightning_count}} - Lightning Strikes (current)</option>
                                <option value="{{"{{"}}lightning_distance}}">{{"{{"}}lightning_distance}} - Lightning Distance km (current)</option>
                                <option value="{{"{{"}}last_temperature}}">{{"{{"}}last_temperature}} - Temperature °C (previous)</option>
                                <option value="{{"{{"}}last_humidity}}">{{"{{"}}last_humidity}} - Humidity % (previous)</option>
                                <option value="{{"{{"}}last_pressure}}">{{"{{"}}last_pressure}} - Pressure mb (previous)</option>
                                <option value="{{"{{"}}last_wind_speed}}">{{"{{"}}last_wind_speed}} - Wind Speed m/s (previous)</option>
                                <option value="{{"{{"}}last_wind_gust}}">{{"{{"}}last_wind_gust}} - Wind Gust m/s (previous)</option>
                                <option value="{{"{{"}}last_wind_direction}}">{{"{{"}}last_wind_direction}} - Wind Direction° (previous)</option>
                                <option value="{{"{{"}}last_lux}}">{{"{{"}}last_lux}} - Light Lux (previous)</option>
                                <option value="{{"{{"}}last_uv}}">{{"{{"}}last_uv}} - UV Index (previous)</option>
                                <option value="{{"{{"}}last_rain_rate}}">{{"{{"}}last_rain_rate}} - Rain Rate mm (previous)</option>
                                <option value="{{"{{"}}last_lightning_count}}">{{"{{"}}last_lightning_count}} - Lightning Strikes (previous)</option>
                                <option value="{{"{{"}}last_lightning_distance}}">{{"{{"}}last_lightning_distance}} - Lightning Distance km (previous)</option>
                            </select>
                        </div>
                        <input type="text" id="smsTo" placeholder="Phone number(s) (comma-separated)" />
                        <textarea id="smsMessage" rows="3" placeholder="SMS message (keep it short)..."></textarea>
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
                                   placeholder="Search or add new tag..." 
                                   autocomplete="off" />
                            <div id="tagDropdown" class="tag-dropdown"></div>
                        </div>
                    </div>
                    <small>Click to select existing tags or type to create new ones</small>
                </div>
                
                <div class="form-group">
                    <label>Cooldown (seconds)</label>
                    <input type="number" id="alarmCooldown" value="1800" />
                </div>
                
                <div class="form-group checkbox-group">
                    <input type="checkbox" id="alarmEnabled" checked />
                    <label for="alarmEnabled">Enabled</label>
                </div>
                
                <div class="modal-actions">
                    <button type="button" class="btn btn-secondary" onclick="closeModal()">Cancel</button>
                    <button type="submit" class="btn btn-primary">Save Alarm</button>
                </div>
            </form>
        </div>
    </div>
    
    <div id="notification" class="notification"></div>
    
    <script src="/alarm-editor/static/script.js"></script>
</body>
</html>`
