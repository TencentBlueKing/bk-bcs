# -*- coding: utf-8 -*-
import pytest

from backend.helm.toolkit import chart_versions

pytestmark = pytest.mark.django_db


def test_snapshot_to_version():
    snapshot = chart_versions.ChartVersionSnapshot(
        **{
            "version": "0.1.0",
            "digest": "123",
            "name": "test",
            "home": "",
            "description": "",
            "engine": "test",
            "created": "2021-01-15 10:00:00",
            "maintainers": '["admin"]',
            "sources": '[]',
            "urls": '["http://repo.example.com/test/test/0.1.0.tgz"]',
            "files": '{}',
            "questions": '{}',
        }
    )
    chart = chart_versions.Chart(name="test")
    chart_version = chart_versions.release_snapshot_to_version(snapshot, chart)
    assert chart_version.chart.name == chart.name
    assert chart_version.version == snapshot.version


@pytest.mark.parametrize(
    "versions, sorted_versions",
    [
        (
            [
                {"version": "0.1.10", "created": "2021-01-15 10:00:00"},
                {"version": "0.1.11", "created": "2021-01-15 11:00:00"},
            ],
            [
                {"version": "0.1.11", "created": "2021-01-15 11:00:00"},
                {"version": "0.1.10", "created": "2021-01-15 10:00:00"},
            ],
        ),
        (
            [
                {"version": "0.1.10", "created": "2021-01-15 10:00:00"},
                {"version": "0.1.11", "created": "2021-01-15 10:00:00"},
            ],
            [
                {"version": "0.1.10", "created": "2021-01-15 10:00:00"},
                {"version": "0.1.11", "created": "2021-01-15 10:00:00"},
            ],
        ),
    ],
)
def test_sort_version_list(versions, sorted_versions):
    assert chart_versions.sort_version_list(versions) == sorted_versions
