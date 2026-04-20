"""Unit tests for fraud detection rule engine."""

import time
from datetime import datetime, timedelta, timezone
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from app.engine.amount import check_amount_anomaly, compute_new_average
from app.engine.device import check_device_change
from app.engine.dormant import check_dormant_account
from app.engine.geo import check_geo_anomaly, haversine


# ─── Amount Anomaly ───

class TestAmountAnomaly:
    def test_insufficient_history(self):
        result = check_amount_anomaly(amount=100000, avg_amount=50000, tx_count=2)
        assert result["triggered"] is False
        assert result["reason"] == "insufficient_history"

    def test_normal_amount(self):
        result = check_amount_anomaly(amount=60000, avg_amount=50000, tx_count=10)
        assert result["triggered"] is False

    def test_anomalous_amount(self):
        result = check_amount_anomaly(amount=200000, avg_amount=50000, tx_count=10, std_dev_multiplier=3.0)
        assert result["triggered"] is True
        assert result["threshold"] == 150000

    def test_exact_threshold(self):
        result = check_amount_anomaly(amount=150000, avg_amount=50000, tx_count=10, std_dev_multiplier=3.0)
        assert result["triggered"] is False  # not > threshold, equal

    def test_compute_new_average(self):
        # (50000 * 10 + 60000) / 11 = 50909.09...
        new_avg = compute_new_average(50000.0, 10, 60000)
        assert abs(new_avg - 50909.09) < 1

    def test_compute_new_average_first_tx(self):
        new_avg = compute_new_average(0.0, 0, 100000)
        assert new_avg == 100000.0


# ─── Device Change ───

class TestDeviceChange:
    def test_no_previous_device(self):
        result = check_device_change("device-123", None)
        assert result["triggered"] is False
        assert result["reason"] == "no_previous_device"

    def test_no_current_device(self):
        result = check_device_change("", "device-old")
        assert result["triggered"] is False
        assert result["reason"] == "no_current_device"

    def test_same_device(self):
        result = check_device_change("device-123", "device-123")
        assert result["triggered"] is False

    def test_different_device(self):
        result = check_device_change("device-new", "device-old")
        assert result["triggered"] is True


# ─── Dormant Account ───

class TestDormantAccount:
    def test_no_prior_transaction(self):
        result = check_dormant_account(None, dormant_days=90)
        assert result["triggered"] is False
        assert result["reason"] == "no_prior_transaction"

    def test_active_account(self):
        last_tx = datetime.now(timezone.utc) - timedelta(days=10)
        result = check_dormant_account(last_tx, dormant_days=90)
        assert result["triggered"] is False
        assert result["days_inactive"] == 10

    def test_dormant_account(self):
        last_tx = datetime.now(timezone.utc) - timedelta(days=100)
        result = check_dormant_account(last_tx, dormant_days=90)
        assert result["triggered"] is True
        assert result["days_inactive"] == 100

    def test_exactly_at_threshold(self):
        last_tx = datetime.now(timezone.utc) - timedelta(days=90)
        result = check_dormant_account(last_tx, dormant_days=90)
        assert result["triggered"] is True


# ─── Geolocation ───

class TestGeoAnomaly:
    def test_no_current_location(self):
        result = check_geo_anomaly(None, None, -6.2, 106.8, datetime.now(timezone.utc))
        assert result["triggered"] is False
        assert result["reason"] == "no_current_location"

    def test_no_previous_location(self):
        result = check_geo_anomaly(-6.2, 106.8, None, None, datetime.now(timezone.utc))
        assert result["triggered"] is False
        assert result["reason"] == "no_previous_location"

    def test_normal_distance(self):
        # Jakarta to Bandung ~150km
        last_tx = datetime.now(timezone.utc) - timedelta(minutes=30)
        result = check_geo_anomaly(-6.9, 107.6, -6.2, 106.8, last_tx, max_distance_km=500)
        assert result["triggered"] is False

    def test_impossible_travel(self):
        # Jakarta to Surabaya ~700km in 30 minutes
        last_tx = datetime.now(timezone.utc) - timedelta(minutes=30)
        result = check_geo_anomaly(-7.25, 112.75, -6.2, 106.8, last_tx, max_distance_km=500)
        assert result["triggered"] is True
        assert result["distance_km"] > 500

    def test_outside_time_window(self):
        # Far distance but 2 hours ago — outside 1hr window
        last_tx = datetime.now(timezone.utc) - timedelta(hours=2)
        result = check_geo_anomaly(-7.25, 112.75, -6.2, 106.8, last_tx,
                                   max_distance_km=500, time_window_hours=1.0)
        assert result["triggered"] is False
        assert result["reason"] == "outside_time_window"

    def test_haversine_known_distance(self):
        # Jakarta (-6.2, 106.8) to Singapore (1.35, 103.82) ≈ 885 km
        dist = haversine(-6.2, 106.8, 1.35, 103.82)
        assert 880 < dist < 900
