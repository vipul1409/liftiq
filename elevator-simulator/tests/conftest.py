"""
conftest.py
Shared pytest fixtures for the elevator simulator test suite.
"""
import sys
import os

# Ensure the elevator-simulator package root is on the path so tests can
# import elevator_state, api, etc. without installation.
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))
