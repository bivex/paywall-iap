#!/usr/bin/env python3
"""
StarUML Wireframe to Markdown Parser

This script parses StarUML project files (.mdj) and extracts wireframe diagrams
into structured Markdown documentation.

Usage:
    python uml_wireframe_parser.py <input.mdj> [output.md]
"""

import json
import sys
import re
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple
from dataclasses import dataclass, field


@dataclass
class WidgetInfo:
    """Information about a wireframe widget."""
    id: str
    type: str
    label: str
    text: str = ""
    documentation: str = ""
    x: int = 0
    y: int = 0
    width: int = 0
    height: int = 0
    parent_id: str = ""

    @property
    def display_label(self) -> str:
        """Get the display label for the widget."""
        return self.label or self.text or "-"

    @property
    def sort_key(self) -> Tuple[int, int]:
        """Sort key for layout positioning (y, then x)."""
        return (self.y, self.x)


@dataclass
class DiagramInfo:
    """Information about a wireframe diagram."""
    id: str
    name: str
    parent_module: str = ""
    documentation: str = ""
    widgets: List[WidgetInfo] = field(default_factory=list)


# Widget type mapping for human-readable names
WIDGET_TYPE_MAP = {
    # UMLWF prefix (standard StarUML)
    "UMLWFButton": "Button",
    "UMLWFTextField": "Text Field",
    "UMLWFTextArea": "Text Area",
    "UMLWFLabel": "Label",
    "UMLWFCheckBox": "Checkbox",
    "UMLWFCheckBoxGroup": "Checkbox Group",
    "UMLWFRadioButton": "Radio Button",
    "UMLWFImage": "Image",
    "UMLWFPanel": "Panel",
    "UMLWFWindow": "Window",
    "UMLWFTextBox": "Text Box",
    "UMLWFLink": "Link",
    "UMLWFTable": "Table",
    "UMLWFTableRow": "Table Row",
    "UMLWFTableColumn": "Table Column",
    "UMLWFComboBox": "Combo Box",
    "UMLWFListBox": "List Box",
    "UMLWFSlider": "Slider",
    "UMLWFProgressBar": "Progress Bar",
    "UMLWFSpinner": "Spinner",
    "UMLWFDatePicker": "Date Picker",
    "UMLWFCanvas": "Canvas",
    "UMLWFLine": "Line",
    "UMLWFEmbed": "Embed",
    "UMLWFIcon": "Icon",
    "UMLWFNavigationBar": "Navigation Bar",
    "UMLWFTabBar": "Tab Bar",
    "UMLWFTreeView": "Tree View",
    "UMLWFMenu": "Menu",
    "UMLWFContextMenu": "Context Menu",
    "UMLWFWebView": "Web View",
    "UMLWFMapView": "Map View",
    "UMLWFAdView": "Ad View",
    "UMLWFSearchBar": "Search Bar",
    "UMLWFScrollView": "Scroll View",
    "UMLWFSegment": "Segment",
    "UMLWFActivityIndicator": "Activity Indicator",
    "UMLWFStepper": "Stepper",
    "UMLWFSwitch": "Switch",
    "UMLWFPageControl": "Page Control",
    # WF prefix (Wireframe plugin)
    "WFButtonView": "Button",
    "WFTextView": "Text",
    "WFInputView": "Input",
    "WFCheckboxView": "Checkbox",
    "WFLinkView": "Link",
    "WFWebFrameView": "Web Frame",
    "WFDesktopFrameView": "Desktop Frame",
    "WFPanelView": "Panel",
    "WFSeparatorView": "Separator",
    "WFAvatarView": "Avatar",
    "WFImageView": "Image",
    "WFTextAreaView": "Text Area",
    "WFSelectView": "Select",
    "WFRadioView": "Radio",
    "WFTableView": "Table",
    "WFListView": "List",
    "WFCardView": "Card",
    "WFHeaderView": "Header",
    "WFFooterView": "Footer",
    "WFNavigationView": "Navigation",
    "WFTabView": "Tab Bar",
    "WFProgressBarView": "Progress Bar",
    "WFBadgeView": "Badge",
    "WFTagView": "Tag",
    "WFAlertView": "Alert",
    "WFModalView": "Modal",
    "WFPopupView": "Popup",
    "WFDropdownView": "Dropdown",
    "WFAccordionView": "Accordion",
    "WFCarouselView": "Carousel",
    "WFTimelineView": "Timeline",
    "WFChartView": "Chart",
    "WFCalendarView": "Calendar",
    "WFMapView": "Map",
    "WFVideoView": "Video",
    "WFAudioView": "Audio",
    "WFFileView": "File",
    "WFFolderView": "Folder",
    "WFBreadcrumbView": "Breadcrumb",
    "WFPaginationView": "Pagination",
    "WFSearchView": "Search",
    "WFFilterView": "Filter",
    "WFSortView": "Sort",
    "WFMenuView": "Menu",
    "WFContextMenuView": "Context Menu",
    "WFTooltipView": "Tooltip",
    "WFSpinnerView": "Spinner",
    "WFSwitchView": "Switch",
    "WFSliderView": "Slider",
    "WFStepperView": "Stepper",
    "WFRatingView": "Rating",
    "WFProgressBarCircularView": "Circular Progress",
    "WFToggleButtonView": "Toggle Button",
    "WFIconButtonView": "Icon Button",
    "WFSplitView": "Split",
    "WFGroupBoxView": "Group Box",
    "WFTabViewView": "Tab View",
    "WFScrollView": "Scroll View",
    "WFCollectionView": "Collection",
    "WFGridView": "Grid",
    "WFDataTableView": "Data Table",
    "WFTreeView": "Tree View",
    "WFMenuItemView": "Menu Item",
    "WFAccordionItemView": "Accordion Item",
    "WFTimelineItemView": "Timeline Item",
    "WFListItemView": "List Item",
    "WFCardItemView": "Card Item",
    "WFTableItemView": "Table Item",
    "WFGridItemView": "Grid Item",
    "WFCollectionItemView": "Collection Item",
}


def get_widget_type_name(widget_type: str) -> str:
    """Get human-readable widget type name."""
    if widget_type in WIDGET_TYPE_MAP:
        return WIDGET_TYPE_MAP[widget_type]
    # Remove common prefixes and suffixes
    result = widget_type
    for prefix in ("UMLWF", "WF"):
        if result.startswith(prefix):
            result = result[len(prefix):]
    # Remove "View" suffix
    if result.endswith("View"):
        result = result[:-4]
    return result


def sanitize_markdown(text: str) -> str:
    """Sanitize text for Markdown output."""
    if not text:
        return ""
    # Escape special Markdown characters
    text = text.replace("|", "\\|")
    text = text.replace("\n", " ")
    return text


def truncate_text(text: str, max_length: int = 50) -> str:
    """Truncate text to a maximum length."""
    if not text:
        return ""
    if len(text) <= max_length:
        return text
    return text[:max_length-3] + "..."


def generate_element_id(widget: WidgetInfo, index: int) -> str:
    """Generate a readable element ID from widget type and index."""
    type_short = widget.type
    # Remove common prefixes
    for prefix in ("UMLWF", "WF"):
        if type_short.startswith(prefix):
            type_short = type_short[len(prefix):]
    # Remove "View" suffix
    if type_short.endswith("View"):
        type_short = type_short[:-4]
    return f"{type_short.lower()}_{index:03d}"


class StarUMLWireframeParser:
    """Parser for StarUML wireframe diagrams."""

    def __init__(self, data: Dict[str, Any]):
        """Initialize parser with StarUML JSON data."""
        self.data = data
        self.element_map: Dict[str, Dict[str, Any]] = {}
        self._build_element_map()

    def _build_element_map(self) -> None:
        """Build a map of all elements by their ID for quick lookup."""
        def traverse(elements: List[Dict[str, Any]]) -> None:
            for elem in elements:
                elem_id = elem.get("_id")
                if elem_id:
                    self.element_map[elem_id] = elem
                if "ownedElements" in elem:
                    traverse(elem["ownedElements"])

        if "ownedElements" in self.data:
            traverse(self.data["ownedElements"])

    def find_element_by_id(self, elem_id: str) -> Optional[Dict[str, Any]]:
        """Find an element by its ID."""
        return self.element_map.get(elem_id)

    def extract_wireframe_diagrams(self) -> List[DiagramInfo]:
        """Extract all wireframe diagrams from the project."""
        diagrams = []

        def traverse(elements: List[Dict[str, Any]], parent_name: str = "") -> None:
            for elem in elements:
                elem_type = elem.get("_type", "")
                elem_name = elem.get("name", "")

                if elem_type in ("UMLWireframeDiagram", "WFWireframeDiagram"):
                    diagram = self._extract_diagram(elem, parent_name)
                    if diagram:
                        diagrams.append(diagram)
                elif elem_type in ("UMLModel", "UMLPackage", "WFWireframe"):
                    new_parent = f"{parent_name}/{elem_name}" if parent_name else elem_name
                    if "ownedElements" in elem:
                        traverse(elem["ownedElements"], new_parent)
                elif "ownedElements" in elem:
                    traverse(elem["ownedElements"], parent_name)

        if "ownedElements" in self.data:
            traverse(self.data["ownedElements"])

        return diagrams

    def _extract_diagram(self, diagram_elem: Dict[str, Any], parent_module: str) -> Optional[DiagramInfo]:
        """Extract information from a wireframe diagram element."""
        diagram_id = diagram_elem.get("_id", "")
        name = diagram_elem.get("name", "Unnamed Diagram")
        documentation = diagram_elem.get("documentation", "")

        diagram = DiagramInfo(
            id=diagram_id,
            name=name,
            parent_module=parent_module,
            documentation=documentation
        )

        # Extract widgets from the diagram
        widgets = self._extract_widgets(diagram_elem)
        diagram.widgets.extend(widgets)

        return diagram

    def _extract_widgets(self, diagram_elem: Dict[str, Any]) -> List[WidgetInfo]:
        """Extract all widgets from a diagram."""
        widgets = []

        # Look for views in the diagram
        if "ownedViews" in diagram_elem:
            for view in diagram_elem["ownedViews"]:
                widget = self._extract_widget_from_view(view)
                if widget:
                    widgets.append(widget)

        return widgets

    def _extract_widget_from_view(self, view: Dict[str, Any]) -> Optional[WidgetInfo]:
        """Extract widget information from a view element."""
        view_type = view.get("_type", "")
        view_id = view.get("_id", "")

        # Check if this is a widget view (support both UMLWF and WF prefixes)
        if not (view_type.startswith("UMLWF") or view_type.startswith("WF")):
            return None

        # Get the model element for additional information
        model_ref = view.get("model", {})
        model_id = model_ref.get("$ref") if isinstance(model_ref, dict) else model_ref

        model_elem = self.find_element_by_id(model_id) if model_id else None

        # Extract text from subViews (LabelView typically contains the actual text)
        text_from_subview = ""
        if "subViews" in view:
            for sub_view in view["subViews"]:
                if sub_view.get("_type") == "LabelView":
                    text_from_subview = sub_view.get("text", "")
                    break

        # Extract basic properties
        widget = WidgetInfo(
            id=view_id,
            type=view_type,
            label=text_from_subview or view.get("text", ""),
            x=int(view.get("left", 0)),
            y=int(view.get("top", 0)),
            width=int(view.get("width", 0)),
            height=int(view.get("height", 0))
        )

        # Extract model properties
        if model_elem:
            if not widget.label:
                widget.label = model_elem.get("name", "")
            widget.documentation = model_elem.get("documentation", "")
            # Check for text property in model
            if "text" in model_elem and not widget.text:
                widget.text = model_elem["text"]

        # Extract additional view properties
        if "documentation" in view and not widget.documentation:
            widget.documentation = view["documentation"]

        # Handle parent reference
        parent_ref = view.get("parent", {})
        if isinstance(parent_ref, dict) and "$ref" in parent_ref:
            widget.parent_id = parent_ref["$ref"]

        # Get name from nameLabel if available
        if "nameLabel" in view:
            name_label_ref = view["nameLabel"]
            if isinstance(name_label_ref, dict) and "$ref" in name_label_ref:
                label_view = self.find_element_by_id(name_label_ref["$ref"])
                if label_view and "text" in label_view:
                    widget.label = label_view["text"]

        return widget


class MarkdownGenerator:
    """Generate Markdown documentation from wireframe diagrams."""

    def __init__(self, diagrams: List[DiagramInfo]):
        """Initialize generator with diagram information."""
        self.diagrams = diagrams

    def generate(self) -> str:
        """Generate complete Markdown documentation."""
        if not self.diagrams:
            return "# Wireframe Analysis\n\nNo wireframe diagrams found in the project.\n"

        lines = []

        # Title and overview
        lines.append("# StarUML Wireframe Documentation")
        lines.append("")
        lines.append(f"This document contains **{len(self.diagrams)}** wireframe diagram(s) extracted from the StarUML project.")
        lines.append("")
        lines.append("---")
        lines.append("")

        # Table of contents
        lines.append("## Table of Contents")
        lines.append("")
        for i, diagram in enumerate(self.diagrams, 1):
            anchor = self._create_anchor(diagram.name)
            lines.append(f"{i}. [{diagram.name}](#{anchor})")
        lines.append("")
        lines.append("---")
        lines.append("")

        # Generate each diagram section
        for i, diagram in enumerate(self.diagrams, 1):
            lines.extend(self._generate_diagram_section(diagram, i))
            lines.append("")

        return "\n".join(lines)

    def _create_anchor(self, text: str) -> str:
        """Create a Markdown anchor from text."""
        return text.lower().replace(" ", "-").replace("/", "-").replace("(", "").replace(")", "")

    def _generate_diagram_section(self, diagram: DiagramInfo, index: int) -> List[str]:
        """Generate Markdown section for a single diagram."""
        lines = []

        # Section header
        lines.append(f"## {index}. {diagram.name}")
        lines.append("")

        # Overview
        lines.append("### 1. Overview")
        lines.append("")
        if diagram.parent_module:
            lines.append(f"**Parent Module:** `{diagram.parent_module}`")
            lines.append("")
        if diagram.documentation:
            lines.append(f"**Description:**")
            lines.append("")
            lines.append(diagram.documentation)
            lines.append("")
        lines.append(f"**Element Count:** {len(diagram.widgets)} widgets")
        lines.append("")

        # UI Elements Table
        lines.append("### 2. UI Elements Table")
        lines.append("")
        lines.append("| Element ID | Type | Label/Text | Position | Size | Notes |")
        lines.append("| :--- | :--- | :--- | :--- | :--- | :--- |")

        for widget in diagram.widgets:
            elem_id = generate_element_id(widget, diagram.widgets.index(widget) + 1)
            type_name = get_widget_type_name(widget.type)
            label = sanitize_markdown(truncate_text(widget.display_label, 30))
            position = f"({widget.x}, {widget.y})"
            size = f"{widget.width}×{widget.height}"
            notes = sanitize_markdown(truncate_text(widget.documentation, 40))

            lines.append(f"| {elem_id} | {type_name} | {label} | {position} | {size} | {notes} |")

        lines.append("")

        # Elements by Type Summary
        lines.append("### 3. Elements by Type")
        lines.append("")

        type_counts = {}
        for widget in diagram.widgets:
            type_counts[widget.type] = type_counts.get(widget.type, 0) + 1

        for widget_type, count in sorted(type_counts.items()):
            type_name = get_widget_type_name(widget_type)
            lines.append(f"- **{type_name}**: {count}")
        lines.append("")

        # Layout Positioning
        lines.append("### 4. Layout Positioning")
        lines.append("")
        lines.append("Elements sorted by visual flow (top-to-bottom, left-to-right):")
        lines.append("")

        sorted_widgets = sorted(diagram.widgets, key=lambda w: w.sort_key)
        for i, widget in enumerate(sorted_widgets, 1):
            elem_id = generate_element_id(widget, diagram.widgets.index(widget) + 1)
            type_name = get_widget_type_name(widget.type)
            label = sanitize_markdown(widget.display_label)

            indent = "  " * (widget.x // 20)  # Visual indentation based on x position
            lines.append(f"{i}. {indent}**[{type_name}]** {label}")
            if widget.documentation:
                lines.append(f"   {indent}_{widget.documentation}_")
            lines.append("")

        return lines


def parse_mdj_file(file_path: str) -> Dict[str, Any]:
    """Parse a StarUML .mdj file and return the JSON data."""
    path = Path(file_path)

    if not path.exists():
        raise FileNotFoundError(f"File not found: {file_path}")

    if not path.suffix.lower() == ".mdj":
        raise ValueError(f"Expected .mdj file, got: {path.suffix}")

    with open(path, "r", encoding="utf-8") as f:
        try:
            return json.load(f)
        except json.JSONDecodeError as e:
            raise ValueError(f"Invalid JSON in file: {e}")


def main():
    """Main entry point for the CLI."""
    if len(sys.argv) < 2:
        print("Usage: python uml_wireframe_parser.py <input.mdj> [output.md]")
        print("")
        print("Arguments:")
        print("  input.mdj   Path to the StarUML project file")
        print("  output.md   Optional path for the output Markdown file")
        print("              (defaults to input filename with .md extension)")
        sys.exit(1)

    input_file = sys.argv[1]

    # Determine output file path
    if len(sys.argv) >= 3:
        output_file = sys.argv[2]
    else:
        output_file = str(Path(input_file).with_suffix(".md"))

    try:
        # Parse the input file
        print(f"Parsing {input_file}...")
        data = parse_mdj_file(input_file)

        # Extract wireframe diagrams
        print("Extracting wireframe diagrams...")
        parser = StarUMLWireframeParser(data)
        diagrams = parser.extract_wireframe_diagrams()

        print(f"Found {len(diagrams)} wireframe diagram(s)")

        # Generate Markdown
        print("Generating Markdown documentation...")
        generator = MarkdownGenerator(diagrams)
        markdown = generator.generate()

        # Write output
        output_path = Path(output_file)
        output_path.write_text(markdown, encoding="utf-8")

        print(f"")
        print(f"Success! Documentation written to: {output_file}")
        print(f"Total diagrams processed: {len(diagrams)}")

        # Print summary
        if diagrams:
            print(f"")
            print("Diagrams found:")
            for diagram in diagrams:
                widget_count = len(diagram.widgets)
                print(f"  - {diagram.name}: {widget_count} widgets")

    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()