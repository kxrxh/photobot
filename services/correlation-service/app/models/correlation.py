from enum import StrEnum

from pydantic import BaseModel, Field


class ParameterGroup(StrEnum):
    """Enum for parameter groups in correlation analysis"""

    COLOR = "color"
    GEOMETRY = "geometry"
    MEDIAN = "median"
    ALL = "all"


class ObjectGroup(BaseModel):
    """Model for a group of objects in correlation analysis"""

    name: str = Field(description="Name of the group")
    object_ids: list[int] = Field(description="List of object IDs in the group")


class CorrelationRequest(BaseModel):
    """Request model for correlation analysis"""

    fractions: list[ObjectGroup] = Field(description="List of object groups to analyze")
    parameter_groups: list[ParameterGroup] = Field(
        description="List of parameter groups to include in analysis",
        default=[ParameterGroup.ALL],
    )


class Condition(BaseModel):
    """Model for a condition in correlation analysis"""

    attribute: str = Field(description="Attribute name")
    operator: str = Field(description="Comparison operator (> or <=)")
    value: float = Field(description="Threshold value")

    model_config = {"frozen": True}  # Make conditions hashable for sets


class CorrelationBase(BaseModel):
    """Base model for correlation results"""

    name: str = Field(description="Name of the group")
    conditions: list[Condition] = Field(description="List of conditions for the group")

    model_config = {"from_attributes": True}


class ConditionTestResult(BaseModel):
    """Result of testing a condition on the input data"""

    true_positives: int = Field(description="Number of true positives")
    false_positives: int = Field(description="Number of false positives")
    true_negatives: int = Field(description="Number of true negatives")
    false_negatives: int = Field(description="Number of false negatives")
    precision: float = Field(description="Precision score")
    recall: float = Field(description="Recall score")
    accuracy: float = Field(description="Accuracy score")
    f1_score: float = Field(description="F1 score")


class CorrelationWithTest(CorrelationBase):
    """Extended correlation model with test results"""

    test_results: ConditionTestResult | None = Field(
        description="Test results for the conditions", default=None
    )

    model_config = {"from_attributes": True}
