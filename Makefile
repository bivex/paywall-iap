.PHONY: dump-routes

ROUTES_OUT ?= routes.txt

dump-routes:
	@echo "Dumping API routes to $(ROUTES_OUT)"
	@$(MAKE) --no-print-directory -C backend dump-routes > $(ROUTES_OUT)
	@echo "Done: $(ROUTES_OUT)"