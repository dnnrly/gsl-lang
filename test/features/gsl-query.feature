Feature: gsl-query CLI validation
  As a developer using gsl-query
  I want the CLI to behave correctly for basic operations
  So that I can trust it to query and transform GSL graphs

  Background:
    Given the gsl-query binary is available

  Scenario: Prints help with -h flag
    When I run gsl-query with parameters "-h"
    Then the app exits without error
    And the app output contains "Usage:"

  Scenario: Prints version information
    When I run gsl-query with parameters "version"
    Then the app exits without error
    And the app output contains "gsl-query version"
    And the app output contains "Commit:"
    And the app output contains "Go:"

  Scenario: Runs a subgraph query
    Given a GSL input file with content:
      """
      set payments
      node PaymentSvc @payments
      node LegacySvc
      PaymentSvc -> LegacySvc
      """
    When I run gsl-query with query "subgraph in payments" using input file
    Then the app exits without error
    And the app output contains "PaymentSvc"

  Scenario: Runs a pipeline query
    Given a GSL input file with content:
      """
      node API
      node DB
      node Cache
      API -> DB
      API -> Cache
      """
    When I run gsl-query with query "remove orphans" using input file
    Then the app exits without error
    And the app output contains "API"
    And the app output contains "DB"
    And the app output contains "Cache"

  Scenario: Errors on invalid query syntax
    Given a GSL input file with content:
      """
      node API
      """
    When I run gsl-query with parameters "--input <input_file> invalid query syntax !!!!"
    Then the app exits with an error

  Scenario: Errors on non-existent input file
    When I run gsl-query with parameters "subgraph exists"
    Then the app exits with an error
