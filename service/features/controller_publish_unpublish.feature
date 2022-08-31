Feature: VxFlex OS CSI interface
  As a consumer of the CSI interface
  I want to test controller publish / unpublish interfaces
  So that they are known to work

@publish-unpublish
  Scenario: Publish volume with single writer
    Given a VxFlexOS service
    And a valid volume
    When I call Probe
    And I call PublishVolume with <access>
    Then a valid PublishVolumeResponse is returned
    And the number of SDC mappings is 1

    Examples:
      | access                      |
      | "single-writer"             |
      | "single-node-single-writer" |
      | "single-node-multi-writer"  |

@publish-unpublish
  Scenario: Publish legacy volume that is on non default array
    Given a VxFlexOS service
    And I induce error "LegacyVolumeConflictError"
    And a valid volume
    When I call Probe
    And I call PublishVolume with <access>
    Then the error contains "Expecting this volume id only on default system.  Aborting operation"

    Examples:
      | access                      |
      | "single-writer"             |
      | "single-node-single-writer" |
      | "single-node-multi-writer"  |

@publish-unpublish
  Scenario: Publish volume but ID is too short to get first 24 bits
    Given a VxFlexOS service
    And a valid volume
    When I call Probe
    And I induce error "VolumeIDTooShortError"
    And I call PublishVolume with <access>
    Then the error contains "is shorter than 3 chars, returning error"

    Examples:
      | access                      |
      | "single-writer"             |
      | "single-node-single-writer" |
      | "single-node-multi-writer"  |

@publish-unpublish
  Scenario: Calling probe twice, so UpdateVolumePrefixToSystemsMap gets a key,value already added
    Given a VxFlexOS service
    And a valid volume
    When I call Probe
    And I call Probe
    Then the error contains "none"

@publish-unpublish
  Scenario Outline: Publish Volume with Wrong Access Types
    Given a VxFlexOS service
    And a valid volume
    And I use AccessType Mount
    When I call Probe
    And I call PublishVolume with <access>
    Then the error contains <msg>

    Examples:
      | access                | msg                                        |
      | "multiple-writer"     | "Mount multinode multi-writer not allowed" |
      | "multi-single-writer" | "Multinode single writer not supported"    |

@publish-unpublish
  Scenario: Idempotent publish volume with single writer
    Given a VxFlexOS service
    And a valid volume
    When I call Probe
    And I call PublishVolume with <access>
    And I call PublishVolume with <access>
    Then a valid PublishVolumeResponse is returned
    And the number of SDC mappings is 1

    Examples:
      | access                      |
      | "single-writer"             |
      | "single-node-single-writer" |
      | "single-node-multi-writer"  |

@publish-unpublish
  Scenario: Publish block volume with multiple writers to single writer volume
    Given a VxFlexOS service
    And a valid volume
    When I call Probe
    And I call PublishVolume with <access>
    And then I use a different nodeID
    And I call PublishVolume with <access>
    Then the error contains "volume already published"

    Examples:
      | access                      |
      | "single-writer"             |
      | "single-node-single-writer" |
      | "single-node-multi-writer"  |

@publish-unpublish
  Scenario: Publish block volume with multiple writers to multiple writer volume
    Given a VxFlexOS service
    And a valid volume
    When I call Probe
    And I call PublishVolume with "multiple-writer"
    And then I use a different nodeID
    And I call PublishVolume with "multiple-writer"
    Then a valid PublishVolumeResponse is returned
    And the number of SDC mappings is 2

@publish-unpublish
  Scenario: Publish block volume with multiple writers to multiple reader volume
    Given a VxFlexOS service
    And a valid volume
    When I call Probe
    And I call PublishVolume with "multiple-reader"
    And then I use a different nodeID
    And I call PublishVolume with "multiple-reader"
    Then the error contains "not compatible with access type"

@publish-unpublish
  Scenario: Publish mount volume with multiple writers to single writer volume
    Given a VxFlexOS service
    And a valid volume
    And I use AccessType Mount
    When I call Probe
    And I call PublishVolume with <access>
    And then I use a different nodeID
    And I call PublishVolume with <access>
    Then the error contains "volume already published"

    Examples:
      | access                      |
      | "single-writer"             |
      | "single-node-single-writer" |
      | "single-node-multi-writer"  |

@publish-unpublish
  Scenario: Publish mount volume with multiple readers to multiple reader volume
    Given a VxFlexOS service
    And a valid volume
    And I use AccessType Mount
    When I call Probe
    And I call PublishVolume with "multiple-reader"
    And then I use a different nodeID
    And I call PublishVolume with "multiple-reader"
    Then a valid PublishVolumeResponse is returned

@publish-unpublish
  Scenario: Publish mount volume with multiple readers to multiple reader volume
    Given a VxFlexOS service
    And a valid volume
    And I use AccessType Mount
    When I call Probe
    And I call PublishVolume with "multiple-reader"
    And then I use a different nodeID
    And I call PublishVolume with "multiple-reader"
    Then a valid PublishVolumeResponse is returned
    And the number of SDC mappings is 2

@publish-unpublish
  Scenario: Publish volume with an invalid volumeID
    Given a VxFlexOS service
    When I call Probe
    And an invalid volume
    And I call PublishVolume with "single-writer"
    Then the error contains "volume not found"

@publish-unpublish
  Scenario: Publish volume no volumeID specified
    Given a VxFlexOS service
    And no volume
    When I call Probe
    And I call PublishVolume with "single-writer"
    Then the error contains "volume ID is required"

@publish-unpublish
  Scenario: Publish volume with no nodeID specified
    Given a VxFlexOS service
    And a valid volume
    And no node
    When I call Probe
    And I call PublishVolume with "single-writer"
    Then the error contains "node ID is required"

@publish-unpublish
  Scenario: Publish volume with no volume capability
    Given a VxFlexOS service
    And a valid volume
    And no volume capability
    When I call Probe
    And I call PublishVolume with "single-writer"
    Then the error contains "volume capability is required"

@publish-unpublish
  Scenario: Publish volume with no access mode
    Given a VxFlexOS service
    And a valid volume
    And no access mode
    When I call Probe
    And I call PublishVolume with "single-writer"
    Then the error contains "access mode is required"


@publish-unpublish
  Scenario: Publish volume with getSDCID error
    Given a VxFlexOS service
    And a valid volume
    And I induce error "GetSdcInstancesError"
    When I call Probe
    And I call PublishVolume with "single-writer"
    Then the error contains "error finding SDC from GUID"

@publish-unpublish
  Scenario: Publish volume with bad vol ID
    Given a VxFlexOS service
    And a valid volume
    And I induce error "BadVolIDError"
    When I call Probe
    And I call PublishVolume with "single-writer"
    Then the error contains "volume not found"


@publish-unpublish
  Scenario: Publish volume with a map SDC error
    Given a VxFlexOS service
    And a valid volume
    And I induce error "MapSdcError"
    When I call Probe
    And I call PublishVolume with "single-writer"
    Then the error contains "error mapping volume to node"

@publish-unpublish
  Scenario: Publish volume with AccessMode UNKNOWN
    Given a VxFlexOS service
    And a valid volume
    When I call Probe
    And I call PublishVolume with "unknown"
    Then the error contains "access mode cannot be UNKNOWN"

@publish-unpublish
  Scenario: Unpublish volume
    Given a VxFlexOS service
    And a valid volume
    When I call Probe
    And I call PublishVolume with "single-writer"
    And no error was received
    And the number of SDC mappings is 1
    And I call UnpublishVolume
    And no error was received
    Then a valid UnpublishVolumeResponse is returned
    And the number of SDC mappings is 0

@publish-unpublish
  Scenario: Idempotent unpublish volume
    Given a VxFlexOS service
    And a valid volume
    When I call Probe
    And I call PublishVolume with "single-writer"
    And no error was received
    And I call UnpublishVolume
    And no error was received
    And I call UnpublishVolume
    And no error was received
    Then a valid UnpublishVolumeResponse is returned

@publish-unpublish
  Scenario: Unpublish volume with no volume id
    Given a VxFlexOS service
    And a valid volume
    When I call Probe
    And I call PublishVolume with "single-writer"
    And no error was received
    And no volume
    And I call UnpublishVolume
    Then the error contains "volume ID is required"

@publish-unpublish
  Scenario: Unpublish volume with invalid volume id
    Given a VxFlexOS service
    And a valid volume
    When I call Probe
    And I call PublishVolume with "single-writer"
    And no error was received
    And an invalid volume
    And I call UnpublishVolume
    Then the error contains "volume not found"

@publish-unpublish
  Scenario: Unpublish volume with no node id
    Given a VxFlexOS service
    And a valid volume
    When I call Probe
    And I call PublishVolume with "single-writer"
    And no error was received
    And no node
    And I call UnpublishVolume
    Then the error contains "Node ID is required"

@publish-unpublish
  Scenario: Unpublish volume with RemoveMappedSdcError
    Given a VxFlexOS service
    And a valid volume
    When I call Probe
    And I call PublishVolume with "single-writer"
    And no error was received
    And I induce error "RemoveMappedSdcError"
    And I call UnpublishVolume
    Then the error contains "Error unmapping volume from node"

@publish-unpublish
  Scenario: Publish / unpublish mount volume with multiple writers to multiple writer volume
    Given a VxFlexOS service
    And a valid volume
    When I call Probe
    And I call PublishVolume with "multiple-writer"
    And a valid PublishVolumeResponse is returned
    And the number of SDC mappings is 1
    And then I use a different nodeID
    And I call PublishVolume with "multiple-writer"
    And a valid PublishVolumeResponse is returned
    And the number of SDC mappings is 2
    And I call UnpublishVolume
    And no error was received
    And the number of SDC mappings is 1
    And then I use a different nodeID
    And I call UnpublishVolume
    And no error was received
    Then the number of SDC mappings is 0
