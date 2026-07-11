Feature: gsl-diagram CLI validation
  As a developer using gsl-diagram
  I want the CLI to behave correctly for basic operations
  So that I can trust it to convert GSL graphs to diagrams

  Background:
    Given the gsl-diagram binary is available

  Scenario: Prints help with -h flag
    When I run gsl-diagram with parameters "-h"
    Then the app exits without error
    And the app output contains "Usage:"

  Scenario: Prints version information
    When I run gsl-diagram with parameters "version"
    Then the app exits without error
    And the app output contains "gsl-diagram version"
    And the app output contains "Commit:"
    And the app output contains "Go:"

  Scenario: Converts GSL to mermaid format
    Given a GSL input file with content:
      """
      node API: "REST API"
      node DB: "Database"
      API -> DB
      """
    When I run gsl-diagram with parameters "--input <input_file> -f mermaid"
    Then the app exits without error
    And the app output contains "graph"
    And the app output contains "API"
    And the app output contains "DB"

  Scenario: Converts GSL to plantuml format
    Given a GSL input file with content:
      """
      node API: "REST API"
      node DB: "Database"
      API -> DB
      """
    When I run gsl-diagram with parameters "--input <input_file> -f plantuml"
    Then the app exits without error
    And the app output contains "@startuml"
    And the app output contains "@enduml"
    And the app output contains "API"
    And the app output contains "DB"

  Scenario: Errors on non-existent input file
    When I run gsl-diagram with parameters "--input nonexistent.gsl"
    Then the app exits with an error

  Scenario: Errors on invalid GSL input
    Given a GSL input file with content:
      """
      this is not valid GSL {{{{
      """
    When I run gsl-diagram with parameters "--input <input_file>"
    Then the app exits with an error
