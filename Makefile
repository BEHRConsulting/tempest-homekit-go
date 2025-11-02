VENDOR_DIR := pkg/web/static
CHART_VERSION := 4.4.4
CHART_DIST := chart.umd.js

.PHONY: vendor-chartjs
ADAPTER_VERSION := 3.0.0
ADAPTER_DIST := chartjs-adapter-date-fns.bundle.min.js

.PHONY: vendor-chart-adapter
vendor-chart-adapter:
	@echo "Vendoring Chart.js adapter v$(ADAPTER_VERSION) into $(VENDOR_DIR)"
	mkdir -p $(VENDOR_DIR)
	curl -fsSL https://unpkg.com/chartjs-adapter-date-fns@$(ADAPTER_VERSION)/dist/$(ADAPTER_DIST) -o $(VENDOR_DIR)/$(ADAPTER_DIST)
	@echo "Vendored $(VENDOR_DIR)/$(ADAPTER_DIST)"

.PHONY: vendor-chartjs
vendor-chartjs: vendor-chart-adapter
	@echo "Vendoring Chart.js v$(CHART_VERSION) into $(VENDOR_DIR)"
	mkdir -p $(VENDOR_DIR)
	curl -fsSL https://unpkg.com/chart.js@$(CHART_VERSION)/dist/chart.umd.js -o $(VENDOR_DIR)/$(CHART_DIST)
	@echo "Vendored $(VENDOR_DIR)/$(CHART_DIST)"

QRCODE_VERSION := 1.4.4
QRCODE_DIST := qrcode.min.js

.PHONY: vendor-qrcode
vendor-qrcode:
	@echo "Vendoring QRCode generator v$(QRCODE_VERSION) into $(VENDOR_DIR)"
	mkdir -p $(VENDOR_DIR)
	curl -fsSL https://unpkg.com/qrcode-generator@$(QRCODE_VERSION)/qrcode.js -o $(VENDOR_DIR)/$(QRCODE_DIST)
	@echo "Vendored $(VENDOR_DIR)/$(QRCODE_DIST)"

.PHONY: vendor-static
vendor-static: vendor-chartjs vendor-qrcode
	@echo "All static files have been updated successfully"
	@echo "Updated files:"
	@echo "  - $(VENDOR_DIR)/$(CHART_DIST) (Chart.js v$(CHART_VERSION))"
	@echo "  - $(VENDOR_DIR)/$(ADAPTER_DIST) (Chart.js adapter v$(ADAPTER_VERSION))"
	@echo "  - $(VENDOR_DIR)/$(QRCODE_DIST) (QRCode generator v$(QRCODE_VERSION))"
