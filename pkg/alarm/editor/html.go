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
                        <button type="button" class="sensor-field-btn" onclick="insertField('temperature')">temperature</button>
                        <button type="button" class="sensor-field-btn" onclick="insertField('humidity')">humidity</button>
                        <button type="button" class="sensor-field-btn" onclick="insertField('pressure')">pressure</button>
                        <button type="button" class="sensor-field-btn" onclick="insertField('wind_speed')">wind_speed</button>
                        <button type="button" class="sensor-field-btn" onclick="insertField('wind_gust')">wind_gust</button>
                        <button type="button" class="sensor-field-btn" onclick="insertField('wind_direction')">wind_direction</button>
                        <button type="button" class="sensor-field-btn" onclick="insertField('lux')">lux</button>
                        <button type="button" class="sensor-field-btn" onclick="insertField('uv')">uv</button>
                        <button type="button" class="sensor-field-btn" onclick="insertField('rain_rate')">rain_rate</button>
                        <button type="button" class="sensor-field-btn" onclick="insertField('rain_daily')">rain_daily</button>
                        <button type="button" class="sensor-field-btn" onclick="insertField('lightning_count')">lightning_count</button>
                        <button type="button" class="sensor-field-btn" onclick="insertField('lightning_distance')">lightning_distance</button>
                    </div>
                    <textarea id="alarmCondition" required></textarea>
                    <small>Click sensor names above to insert into condition. Supports units: 80F or 26.7C (temp), 25mph or 11.2m/s (wind). Change detection: *field (any change), &gt;field (increase), &lt;field (decrease). Examples: temperature &gt; 85F, *lightning_count (any strike), &gt;rain_rate (rain increasing), &lt;lightning_distance (lightning closer)</small>
                </div>
                
                <div class="form-group">
                    <label>Delivery Methods *</label>
                    <div class="delivery-methods">
                        <label class="delivery-method">
                            <input type="checkbox" id="deliveryConsole" checked />
                            <span>📟 Console</span>
                        </label>
                        <label class="delivery-method">
                            <input type="checkbox" id="deliverySyslog" />
                            <span>📋 Syslog</span>
                        </label>
                        <label class="delivery-method">
                            <input type="checkbox" id="deliveryEventlog" />
                            <span>📊 Event Log</span>
                        </label>
                        <label class="delivery-method">
                            <input type="checkbox" id="deliveryEmail" />
                            <span>✉️ Email</span>
                        </label>
                        <label class="delivery-method">
                            <input type="checkbox" id="deliverySMS" />
                            <span>📱 SMS</span>
                        </label>
                    </div>
                    <small>Select at least one delivery method</small>
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
