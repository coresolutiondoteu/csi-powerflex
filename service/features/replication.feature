Feature: PowerFlex replication
  As a powerflex user, I want to test powerflex replication
  So that replication is known to work

@replication
Scenario Outline: Call GetReplicationCapabilities
  Given a VxFlexOS service
  And I induce error <error>
  When I call GetReplicationCapabilities
  Then the error contains <errormsg>
  And a <valid> replication capabilities structure is returned 
  Examples:
  | error                      | errormsg                          | valid |
  | "none"                     | "none"                            | "true"  | 
  

@replication
Scenario Outline: Create and delete a replicated volume
  Given a VxFlexOS service
  When I call CreateVolume <name>
  And I induce error <error>
  And I call CreateRemoteVolume
  Then the error contains <errormsg>
  And a <valid> remote volume is returned
  Examples:
  | name                     | error                        | errormsg                           | valid    |
  | "sourcevol"              | "none"                       | "none"                             | "true"   | 
  | "sourcevol"              | "NoVolIDError"               | "volume ID is required"            | "false"  |
  | "sourcevol"              | "controller-probe"           | "PodmonControllerProbeError"       | "false"  |
  | "sourcevol"              | "GetVolByIDError"            | "can't query volume"               | "false"  |
  | "sourcevol"              | "PeerMdmError"               | "PeerMdmError"                     | "false"  |
  | "sourcevol"              | "CreateVolumeError"          | "create volume induced error"      | "false"  |
  | "sourcevol"              | "BadVolIDError"              | "failed to provide"                | "false"  |
  | "sourcevol"              | "BadRemoteSystemIDError"     | "systemid or systemname not found" | "false"  |
  | "sourcevol"              | "ProbePrimaryError"          | "PodmonControllerProbeError"       | "false"  |
  | "sourcevol"              | "ProbeSecondaryError"        | "PodmonControllerProbeError"       | "false"  |
