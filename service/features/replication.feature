Feature: PowerFlex replication
  As a powerflex user, I want to test powerflex replication
  So that replication is known to work

@replication
Scenario Outline: Test GetReplicationCapabilities
  Given a VxFlexOS service
  And I use config "replication-config"
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
  And I use config "replication-config"
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
  | "sourcevol"              | "BadRemoteSystemIDError"     | "System 15dbbf5617523655 not found"| "false"  |
  | "sourcevol"              | "ProbePrimaryError"          | "PodmonControllerProbeError"       | "false"  |
  | "sourcevol"              | "ProbeSecondaryError"        | "PodmonControllerProbeError"       | "false"  |


@replication
Scenario Outline: Test CreateStorageProtectionGroup
  Given a VxFlexOS service
  And I use config "replication-config"
  When I call CreateVolume <name>
  And I call CreateRemoteVolume
  And I induce error <error>
  And I call CreateStorageProtectionGroup
  Then the error contains <errormsg>
  And a <valid> remote volume is returned
  Examples:
  | name                     | error                                       | errormsg                                            | valid    |
  | "sourcevol"              | "none"                                      | "none"                                              | "true"   | 
  | "sourcevol"              | "NoVolIDError"                              | "volume ID is required"                             | "false"  |
  | "sourcevol"              | "BadVolIDError"                             | "failed to provide"                                 | "false"  |
  | "sourcevol"              | "EmptyParametersListError"                  | "empty parameters list"                             | "false"  |
  | "sourcevol"              | "controller-probe"                          | "PodmonControllerProbeError"                        | "false"  |
  | "sourcevol"              | "GetVolByIDError"                           | "can't query volume"                                | "false"  |
  | "sourcevol"              | "ReplicationConsistencyGroupError"          | "create rcg induced error"                          | "false"  |
  | "sourcevol"              | "GetReplicationConsistencyGroupsError"      | "could not GET ReplicationConsistencyGroups"       | "false"  |
  | "sourcevol"              | "GetRCGByIdError"                           | "could not GET RCG by ID"                           | "false"  |
  | "sourcevol"              | "ProbePrimaryError"                         | "PodmonControllerProbeError"                        | "false"  |
  | "sourcevol"              | "ProbeSecondaryError"                       | "PodmonControllerProbeError"                        | "false"  |
  | "sourcevol"              | "NoProtectionDomainError"                   | "NoProtectionDomainError"                           | "false"  |
  | "sourcevol"              | "ReplicationPairError"                      | "POST ReplicationPair induced error"                | "false"  |
  | "sourcevol"              | "GetReplicationPairError"                   | "GET ReplicationPair induced error"		             | "false"  |
  | "sourcevol"              | "PeerMdmError"                              | "PeerMdmError"                                      | "false"  |
  | "sourcevol"              | "RemoteReplicationConsistencyGroupError"    | "could not GET Remote ReplicationConsistencyGroup"  | "false"  |
  | "sourcevol"              | "BadRemoteSystem"                           | "couldn't getSystem (remote)"                       | "false"  |
  | "sourcevol"              | "FindVolumeIDError"                         | "can't find volume replicated-sourcevol by name"    | "false"  |
  | "sourcevol"              | "StorageGroupAlreadyExists"                 | "none"                                              | "true"   | 
  | "sourcevol"              | "StorageGroupAlreadyExistsUnretriavable"    | "couldn't find replication consistency group"       | "false"  |
  | "sourcevol"              | "ReplicationPairAlreadyExists"              | "none"                                              | "true"   |
  | "sourcevol"              | "ReplicationPairAlreadyExistsUnretrievable" | "couldn't find replication pair"                    | "false"  |

@replication
Scenario Outline: Test CreateStorageProtectionGroup with arguments
  Given a VxFlexOS service
  And I use config "replication-config"
  When I call CreateVolume <name>
  And I call CreateRemoteVolume
  And I induce error <error>
  And I call CreateStorageProtectionGroup with <group name>, <remote cluster id>, <rpo>
  Then the error contains <errormsg>
  And a <valid> remote volume is returned
  Examples:
  | name          | group name | remote cluster id | rpo  | error  | errormsg | valid    |
  | "sourcevol"   | "rcg-1"    | "cluster-k211"    | "60" | "none" | "none"   | "true"   | 
  | "sourcevol"   | ""         | "cluster-k211"    | "60" | "none" | "none"   | "true"   | 
  | "sourcevol"   | ""         | "self"            | "60" | "none" | "none"   | "true"   | 
  | "sourcevol"   | ""         | "k211-boston"     | "60" | "none" | "none"   | "true"   | 

@replication
Scenario Outline: Test multiple CreateStorageProtectionGroup calls
  Given a VxFlexOS service
  And I use config "replication-config"
  When I call CreateVolume <name1>
  And I call CreateRemoteVolume
  And I call CreateStorageProtectionGroup with <group name>, <remote cluster id>, <rpo>
  When I call CreateVolume <name2>
  And I call CreateRemoteVolume
  And I call CreateStorageProtectionGroup with <group name>, <remote cluster id>, <rpo2>
  Then the error contains <errormsg>
  And a <valid> remote volume is returned
  Examples:
  | name1     | name2     | group name | remote cluster id | rpo  | rpo2   | errormsg | valid    |
  | "1srcVol" | "2srcVol" | ""         | "cluster-k211"    | "60" | "60"   | "none"   | "true"   | 
  | "1srcVol" | "2srcVol" | ""         | "cluster-k211"    | "60" | "120"  | "none"   | "true"   | 

@replication
Scenario Outline: Test DeleteStorageProtectionGroup up to volume
  Given a VxFlexOS service
  And I use config "replication-config"
  When I call CreateVolume <name>
  And I call CreateRemoteVolume
  And I call CreateStorageProtectionGroup
  And I induce error <error>
  And I call DeleteVolume <name>
  Then the error contains <errormsg>
  And a <valid> remote volume is returned
  Examples:
  | name                     | error                                       | errormsg                                           | valid    |
  | "sourcevol"              | "none"                                      | "none"                                             | "true"   | 
  | "sourcevol"              | "NoDeleteReplicationPair"                   | "pairs exist"                                      | "false"  |
  | "sourcevol"              | "ReplicationPairAlreadyExistsUnretrievable" | "error removing replication pair"                  | "false"  |
  | "sourcevol"              | "GetReplicationPairError"                   | "GET ReplicationPair induced error"                | "false"  |

@replication
Scenario Outline: Test DeleteStorageProtectionGroup 
  Given a VxFlexOS service
  And I use config "replication-config"
  When I call CreateVolume <name>
  And I call CreateRemoteVolume
  And I call CreateStorageProtectionGroup
  And I call DeleteVolume <name>
  And I induce error <error>
  And I call DeleteStorageProtectionGroup
  Then the error contains <errormsg>
  And a <valid> remote volume is returned
  Examples:
  | name                     | error                                 | errormsg                                           | valid    |
  | "sourcevol"              | "none"                                | "none"                                             | "true"   | 
  | "sourcevol"              | "GetReplicationPairError"             | "GET ReplicationPair induced error"                | "false"  |
  | "sourcevol"              | "ReplicationGroupAlreadyDeleted"      | "none"                                             | "true"   |
  | "sourcevol"              | "GetRCGByIdError"                     | "could not GET RCG by ID"                          | "false"  |

@replication
Scenario Outline: Test GetStorageProtectionGroupStatus 
  Given a VxFlexOS service
  And I use config "replication-config"
  When I call CreateVolume <name>
  And I call CreateRemoteVolume
  And I call CreateStorageProtectionGroup
  And I induce error <error>
  And I call GetStorageProtectionGroupStatus
  Then the error contains <errormsg>
  And a <valid> remote volume is returned
  Examples:
  | name                     | error                                    | errormsg                                           | valid    |
  | "sourcevol"              | "none"                                   | "none"                                             | "true"   |
  | "sourcevol"              | "GetRCGByIdError"                        | "could not GET RCG by ID"                          | "false"  |
  | "sourcevol"              | "GetReplicationPairError"                | "GET ReplicationPair induced error"                | "false"  |

@replication
Scenario Outline: Test GetStorageProtectionGroupStatus current status
  Given a VxFlexOS service
  And I use config "replication-config"
  When I call CreateVolume <name>
  And I call CreateRemoteVolume
  And I call CreateStorageProtectionGroup
  And I call GetStorageProtectionGroupStatus with state <state> and mode <mode>
  Then the error contains <errormsg>
  Examples:
  | name                     | errormsg   | state       | mode                  |
  | "sourcevol"              | "none"     | "Normal"    | "Consistent"          |
  | "sourcevol"              | "none"     | "Normal"    | "PartiallyConsistent" |
  | "sourcevol"              | "none"     | "Normal"    | "ConsistentPending"   |
  | "sourcevol"              | "none"     | "Normal"    | "Invalid"             |
  | "sourcevol"              | "none"     | "Failover"  | "Consistent"          |
  | "sourcevol"              | "none"     | "Paused"    | "Consistent"          |

@replication
Scenario Outline: Test GetStorageProtectionGroupStatus with new replication pairs
  Given a VxFlexOS service
  And I use config "replication-config"
  When I call CreateVolume <name>
  And I call CreateRemoteVolume
  And I call CreateStorageProtectionGroup
  And I call DeleteVolume <name>
  And I call GetStorageProtectionGroupStatus
  Then the error contains <errormsg>
  And a <valid> remote volume is returned
  Examples:
  | name                     | error     | errormsg                       | valid    |
  | "sourcevol"              | "none"    | "no replication pairs exist"   | "false"  |

@replication
Scenario Outline: Test ExecuteAction
  Given a VxFlexOS service
  And I use config "replication-config"
  When I call CreateVolume <name>
  And I call CreateRemoteVolume
  And I call CreateStorageProtectionGroup
  And I call GetStorageProtectionGroupStatus with state <state> and mode <mode>
  And I induce error <error>
  And I call ExecuteAction <action>
  Then the error contains <errormsg>
  And a <valid> remote volume is returned
  Examples:
  | name                     | error                     | errormsg                            | action              | valid   | state      | mode          |
  | "sourcevol"              | "none"                    | "none"                              | "CreateSnapshot"    | "true"  | "Normal"   | "Consistent"  |
  | "sourcevol"              | "ExecuteActionError"      | "could not execute RCG action"      | "CreateSnapshot"    | "false" | "Normal"   | "Consistent"  |
  | "sourcevol"              | "SnapshotCreationError"   | "RCG snapshot not created"          | "CreateSnapshot"    | "false" | "Normal"   | "Consistent"  |
  | "sourcevol"              | "none"                    | "none"                              | "FailoverRemote"    | "true"  | "Normal"   | "Consistent"  |
  | "sourcevol"              | "ExecuteActionError"      | "could not execute RCG action"      | "FailoverRemote"    | "false" | "Normal"   | "Consistent"  |
  | "sourcevol"              | "none"                    | "none"                              | "UnplannedFailover" | "true"  | "Normal"   | "Consistent"  |
  | "sourcevol"              | "ExecuteActionError"      | "could not execute RCG action"      | "UnplannedFailover" | "false" | "Normal"   | "Consistent"  |
  | "sourcevol"              | "none"                    | "none"                              | "ReprotectLocal"    | "true"  | "Normal"   | "Consistent"  |
  | "sourcevol"              | "ExecuteActionError"      | "could not execute RCG action"      | "ReprotectLocal"    | "false" | "Normal"   | "Consistent"  |
  | "sourcevol"              | "none"                    | "none"                              | "Resume"            | "true"  | "Failover" | "Consistent"  |
  | "sourcevol"              | "none"                    | "none"                              | "Resume"            | "true"  | "Paused"   | "Consistent"  |
  | "sourcevol"              | "ExecuteActionError"      | "could not execute RCG action"      | "Resume"            | "false" | "Failover" | "Consistent"  |
  | "sourcevol"              | "none"                    | "none"                              | "Suspend"           | "true"  | "Normal"   | "Consistent"  |
  | "sourcevol"              | "ExecuteActionError"      | "could not execute RCG action"      | "Suspend"           | "false" | "Normal"   | "Consistent"  |
  | "sourcevol"              | "none"                    | "not match with supported actions"  | "Unknown"           | "false" | "Normal"   | "Consistent"  |