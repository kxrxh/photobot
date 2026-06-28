"""Unit tests for correlation models."""

import pytest
from pydantic import ValidationError

from app.models.correlation import (
    Condition,
    CorrelationRequest,
    ObjectGroup,
    ParameterGroup,
)


class TestParameterGroup:
    """Tests for ParameterGroup enum."""

    def test_has_expected_values(self):
        assert ParameterGroup.COLOR.value == "color"
        assert ParameterGroup.GEOMETRY.value == "geometry"
        assert ParameterGroup.MEDIAN.value == "median"
        assert ParameterGroup.ALL.value == "all"


class TestObjectGroup:
    """Tests for ObjectGroup model."""

    def test_creates_with_valid_data(self):
        g = ObjectGroup(name="test", object_ids=[1, 2, 3])
        assert g.name == "test"
        assert g.object_ids == [1, 2, 3]

    def test_empty_object_ids_allowed(self):
        g = ObjectGroup(name="empty", object_ids=[])
        assert g.object_ids == []


class TestCorrelationRequest:
    """Tests for CorrelationRequest model."""

    def test_creates_with_default_parameter_groups(self):
        req = CorrelationRequest(
            fractions=[ObjectGroup(name="g1", object_ids=[1])],
        )
        assert req.parameter_groups == [ParameterGroup.ALL]

    def test_creates_with_explicit_parameter_groups(self):
        req = CorrelationRequest(
            fractions=[ObjectGroup(name="g1", object_ids=[1])],
            parameter_groups=[ParameterGroup.COLOR, ParameterGroup.GEOMETRY],
        )
        assert len(req.parameter_groups) == 2


class TestCondition:
    """Tests for Condition model."""

    def test_creates_with_valid_data(self):
        c = Condition(attribute="m_h", operator=">=", value=0.5)
        assert c.attribute == "m_h"
        assert c.operator == ">="
        assert c.value == 0.5

    def test_is_hashable(self):
        c = Condition(attribute="m_h", operator=">=", value=0.5)
        s = {c}
        assert c in s

    def test_is_frozen(self):
        c = Condition(attribute="m_h", operator=">=", value=0.5)
        with pytest.raises(ValidationError):
            c.attribute = "other"
