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
