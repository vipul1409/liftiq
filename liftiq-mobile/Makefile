.DEFAULT_GOAL := help

.PHONY: help start mobile ios android web typecheck test test-watch test-coverage install build-ios build-android

help:
	@echo ""
	@echo "LiftIQ Mobile (liftiq-mobile)"
	@echo ""
	@echo "  Development"
	@echo "  make start            Start Expo Metro bundler"
	@echo "  make mobile           Alias for start"
	@echo "  make ios              Open in iOS Simulator"
	@echo "  make android          Open in Android Emulator"
	@echo "  make web              Open in browser"
	@echo ""
	@echo "  First-time dev build (required for voice + camera)"
	@echo "  make build-ios        Build and install iOS dev client"
	@echo "  make build-android    Build and install Android dev client"
	@echo ""
	@echo "  Quality"
	@echo "  make typecheck        Run tsc --noEmit"
	@echo "  make test             Run Jest test suite"
	@echo "  make test-watch       Run Jest in watch mode"
	@echo "  make test-coverage    Run Jest with coverage report"
	@echo ""
	@echo "  Setup"
	@echo "  make install          npm install"
	@echo ""
	@echo "  Environment variables (liftiq-mobile/.env):"
	@echo "  EXPO_PUBLIC_COMPLIANCE_API_URL   default: http://localhost:8080"
	@echo "  EXPO_PUBLIC_REPORT_API_URL        default: http://localhost:8082"
	@echo ""

start:
	npx expo start

mobile: start

ios:
	npx expo start --ios

android:
	npx expo start --android

web:
	npx expo start --web

build-ios:
	npx expo run:ios

build-android:
	npx expo run:android

typecheck:
	npx tsc --noEmit

test:
	npm test

test-watch:
	npm run test:watch

test-coverage:
	npm run test:coverage

install:
	npm install
