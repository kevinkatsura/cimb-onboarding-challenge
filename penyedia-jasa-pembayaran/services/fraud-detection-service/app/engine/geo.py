"""Geolocation anomaly detection.

Detects impossible travel by computing the Haversine distance between
the current and last-known location. If the distance exceeds a threshold
within a given time window, the check is triggered.
"""

import math
import logging
from datetime import datetime, timezone

logger = logging.getLogger("fraud.geo")

EARTH_RADIUS_KM = 6371.0


def haversine(lat1: float, lon1: float, lat2: float, lon2: float) -> float:
    """Compute the great-circle distance between two points on Earth (km)."""
    lat1, lon1, lat2, lon2 = map(math.radians, [lat1, lon1, lat2, lon2])
    dlat = lat2 - lat1
    dlon = lon2 - lon1
    a = math.sin(dlat / 2) ** 2 + math.cos(lat1) * math.cos(lat2) * math.sin(dlon / 2) ** 2
    return 2 * EARTH_RADIUS_KM * math.asin(math.sqrt(a))


def check_geo_anomaly(
    current_lat: float | None,
    current_lon: float | None,
    last_lat: float | None,
    last_lon: float | None,
    last_tx_at: datetime | None,
    max_distance_km: float = 500.0,
    time_window_hours: float = 1.0,
) -> dict:
    """Check for impossible-travel geolocation anomaly.

    Args:
        current_lat/lon: current transaction location
        last_lat/lon: last known transaction location from fraud_account_stats
        last_tx_at: timestamp of last transaction
        max_distance_km: distance threshold in km
        time_window_hours: time window — flag only if last tx was within this window

    Returns:
        dict with 'triggered', 'distance_km', 'max_distance_km', 'hours_elapsed'
    """
    # Missing coordinate data — can't evaluate
    if None in (current_lat, current_lon) or current_lat == 0 and current_lon == 0:
        return {
            "triggered": False,
            "rule_code": "GEO_ANOMALY",
            "reason": "no_current_location",
        }

    if None in (last_lat, last_lon) or (last_lat == 0 and last_lon == 0):
        return {
            "triggered": False,
            "rule_code": "GEO_ANOMALY",
            "reason": "no_previous_location",
        }

    if last_tx_at is None:
        return {
            "triggered": False,
            "rule_code": "GEO_ANOMALY",
            "reason": "no_prior_transaction",
        }

    now = datetime.now(timezone.utc)
    if last_tx_at.tzinfo is None:
        last_tx_at = last_tx_at.replace(tzinfo=timezone.utc)

    hours_elapsed = (now - last_tx_at).total_seconds() / 3600.0

    # Only check if last transaction was within the time window
    if hours_elapsed > time_window_hours:
        return {
            "triggered": False,
            "rule_code": "GEO_ANOMALY",
            "reason": "outside_time_window",
            "hours_elapsed": round(hours_elapsed, 2),
        }

    distance_km = haversine(current_lat, current_lon, last_lat, last_lon)
    triggered = distance_km > max_distance_km

    if triggered:
        logger.warning(
            "Geo anomaly (impossible travel): distance=%.1fkm threshold=%.1fkm hours=%.2f",
            distance_km, max_distance_km, hours_elapsed,
        )

    return {
        "triggered": triggered,
        "rule_code": "GEO_ANOMALY",
        "distance_km": round(distance_km, 2),
        "max_distance_km": max_distance_km,
        "hours_elapsed": round(hours_elapsed, 2),
    }
