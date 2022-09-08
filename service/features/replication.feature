Feature: PowerFlex replication
  As a powerflex user, I want to test powerflex replication
  So that replication is known to work

@replication
Scenario Outline: Test GetReplicationCapabilities
  Given a VxFlexOS service
  And I induce error <error>
  When I call GetReplicationCapabilities
  Then the error contains <errormsg>
  And a <valid> replication capabilities structure is returned 
  Examples:
  | error                      | errormsg                          | valid |
  | "none"                     | "none"                            | "true"  | 
  

@replication
Scenario Outline: Test CreateRemoteVolume
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


@replication
Scenario Outline: Test CreateStorageProtectionGroup
  Given a VxFlexOS service
  When I call CreateVolume <name>
  And I call CreateRemoteVolume
  And I induce error <error>
  And I call CreateStorageProtectionGroup
  Then the error contains <errormsg>
  And a <valid> remote volume is returned
  Examples:
  | name                     | error                                    | errormsg                                            | valid    |
  | "sourcevol"              | "none"                                   | "none"                                              | "true"   | 
  | "sourcevol"              | "NoVolIDError"                           | "volume ID is required"                             | "false"  |
  | "sourcevol"              | "BadVolIDError"                          | "failed to provide"                                 | "false"  |
  | "sourcevol"              | "EmptyParametersListError"               | "empty parameters list"                             | "false"  |
  | "sourcevol"              | "controller-probe"                       | "PodmonControllerProbeError"                        | "false"  |
  | "sourcevol"              | "GetVolByIDError"                        | "can't query volume"                                | "false"  |
  | "sourcevol"              | "ReplicationConsistencyGroupError"       | "create rcg induced error"                          | "false"  |
  | "sourcevol"              | "GetReplicationConsistencyGroupError"    | "could not GET ReplicationConsistencyGroup"         | "false"  |
  | "sourcevol"              | "ProbePrimaryError"                      | "PodmonControllerProbeError"                        | "false"  |
  | "sourcevol"              | "ProbeSecondaryError"                    | "PodmonControllerProbeError"                        | "false"  |
  | "sourcevol"              | "NoProtectionDomainError"                | "NoProtectionDomainError"                           | "false"  |
  | "sourcevol"              | "ReplicationPairError"                   | "POST ReplicationPair induced error"                | "false"  |
  | "sourcevol"              | "GetReplicationPairError"                | "GET ReplicationPair induced error"		      | "false"  |
  | "sourcevol"              | "PeerMdmError"                           | "PeerMdmError"                                      | "false"  |
  | "sourcevol"              | "RemoteReplicationConsistencyGroupError" | "could not GET Remote ReplicationConsistencyGroup"  | "false"  |
  | "sourcevol"              | "RemoteRCGBadNameError"                  | "remote replication consistency group not found"    | "false"  |
  | "sourcevol"              | "NoProtectionDomainError"                | "induced error"                                     | "false"  |
  | "sourcevol"              | "BadRemoteSystem"                        | "couldn't getSystem (remote)"                       | "false"  |
  | "sourcevol"              | "FindVolumeIDError"                      | "can't find volume replicated-sourcevol by name"    | "false"  |

@replication
Scenario Outline: Test CreateStorageProtectionGroup
  Given a VxFlexOS service
  When I call CreateVolume <name>
  And I call CreateRemoteVolume
  And I call CreateStorageProtectionGroup
  And I induce error <error>
  And I call DeleteVolume <name>
  And I call DeleteStorageProtectionGroup
  Then the error contains <errormsg>
  And a <valid> remote volume is returned
  Examples:
  | name                     | error                        | errormsg                                           | valid    |
  | "sourcevol"              | "none"                       | "none"                                             | "true"   | 
  | "sourcevol"              | "RemoveRCGError"             | "error deleting the replication consistency group" | "false"   | 
  | "sourcevol"              | "NoDeleteReplicationPair"    | "pairs exist"                                      | "false" |
