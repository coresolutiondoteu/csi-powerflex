// Copyright © 2019-2022 Dell Inc. or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//      http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package service

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	types "github.com/dell/goscaleio/types/v1"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	codes "google.golang.org/grpc/codes"
)

var (
	debug         bool
	sdcMappings   []types.MappedSdcInfo
	sdcMappingsID string

	stepHandlersErrors struct {
		FindVolumeIDError                         bool
		GetVolByIDError                           bool
		GetStoragePoolsError                      bool
		PodmonFindSdcError                        bool
		PodmonVolumeStatisticsError               bool
		PodmonNoNodeIDError                       bool
		PodmonNoSystemError                       bool
		PodmonNoVolumeNoNodeIDError               bool
		PodmonControllerProbeError                bool
		PodmonNodeProbeError                      bool
		PodmonVolumeError                         bool
		GetSystemSdcError                         bool
		GetSdcInstancesError                      bool
		MapSdcError                               bool
		RemoveMappedSdcError                      bool
		SDCLimitsError                            bool
		SIOGatewayVolumeNotFoundError             bool
		GetStatisticsError                        bool
		CreateSnapshotError                       bool
		RemoveVolumeError                         bool
		VolumeInstancesError                      bool
		BadVolIDError                             bool
		NoCsiVolIDError                           bool
		WrongVolIDError                           bool
		WrongSystemError                          bool
		NoEndpointError                           bool
		NoUserError                               bool
		NoPasswordError                           bool
		NoSysNameError                            bool
		NoAdminError                              bool
		WrongSysNameError                         bool
		NoVolumeIDError                           bool
		SetVolumeSizeError                        bool
		systemNameMatchingError                   bool
		LegacyVolumeConflictError                 bool
		VolumeIDTooShortError                     bool
		EmptyEphemeralID                          bool
		IncorrectEphemeralID                      bool
		TooManyDashesVolIDError                   bool
		CorrectFormatBadCsiVolID                  bool
		EmptySysID                                bool
		VolIDListEmptyError                       bool
		CreateVGSNoNameError                      bool
		CreateVGSNameTooLongError                 bool
		CreateVGSLegacyVol                        bool
		CreateVGSAcrossTwoArrays                  bool
		CreateVGSBadTimeError                     bool
		CreateSplitVGSError                       bool
		BadVolIDJSON                              bool
		BadMountPathError                         bool
		NoMountPathError                          bool
		NoVolIDError                              bool
		NoVolIDSDCError                           bool
		NoVolError                                bool
		PeerMdmError                              bool
		CreateVolumeError                         bool
		BadRemoteSystemIDError                    bool
		NoProtectionDomainError                   bool
		GetReplicationConsistencyGroupsError      bool
		ReplicationConsistencyGroupError          bool
		ReplicationPairError                      bool
		GetReplicationPairError                   bool
		RemoteReplicationConsistencyGroupError    bool
		RemoteRCGBadNameError                     bool
		EmptyParametersListError                  bool
		RemoveRCGError                            bool
		NoDeleteReplicationPair                   bool
		BadRemoteSystem                           bool
		ExecuteActionError                        bool
		StorageGroupAlreadyExists                 bool
		StorageGroupAlreadyExistsUnretriavable    bool
		ReplicationGroupAlreadyDeleted            bool
		ReplicationPairAlreadyExists              bool
		ReplicationPairAlreadyExistsUnretrievable bool
		SnapshotCreationError                     bool
		GetRCGByIdError                           bool
	}
)

// This file contains HTTP handlers for mocking to the ScaleIO API.
// This allows unit testing with a Scale IO but still provides some coverage in the goscaleio library.
var scaleioRouter http.Handler
var testControllerHasNoConnection bool
var count int

// getFileHandler returns an http.Handler that
func getHandler() http.Handler {
	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			log.Printf("handler called: %s %s", r.Method, r.URL)
			if scaleioRouter == nil {
				getRouter().ServeHTTP(w, r)
			}
		})
	log.Printf("Clearing volume caches\n")
	volumeIDToName = make(map[string]string)
	volumeIDToAncestorID = make(map[string]string)
	volumeNameToID = make(map[string]string)
	volumeIDToConsistencyGroupID = make(map[string]string)
	volumeIDToSizeInKB = make(map[string]string)
	volumeIDToReplicationState = make(map[string]string)
	rcgIDToName = make(map[string]string)
	rcgNameToID = make(map[string]string)
	replicationConsistencyGroups = make(map[string]map[string]string)
	replicationPairIDToName = make(map[string]string)
	replicationPairNameToID = make(map[string]string)
	replicationPairIDToSourceVolume = make(map[string]string)
	replicationPairIDToDestinationVolume = make(map[string]string)
	rcgIDtoDestinationVolumes = make(map[string][]string)
	debug = false
	stepHandlersErrors.FindVolumeIDError = false
	stepHandlersErrors.GetVolByIDError = false
	stepHandlersErrors.SIOGatewayVolumeNotFoundError = false
	stepHandlersErrors.GetStoragePoolsError = false
	stepHandlersErrors.PodmonFindSdcError = false
	stepHandlersErrors.PodmonVolumeStatisticsError = false
	stepHandlersErrors.PodmonNoVolumeNoNodeIDError = false
	stepHandlersErrors.PodmonNoNodeIDError = false
	stepHandlersErrors.PodmonNoSystemError = false
	stepHandlersErrors.PodmonControllerProbeError = false
	stepHandlersErrors.PodmonNodeProbeError = false
	stepHandlersErrors.PodmonVolumeError = false
	stepHandlersErrors.GetSdcInstancesError = false
	stepHandlersErrors.MapSdcError = false
	stepHandlersErrors.RemoveMappedSdcError = false
	stepHandlersErrors.SDCLimitsError = false
	stepHandlersErrors.GetStatisticsError = false
	stepHandlersErrors.GetSystemSdcError = false
	stepHandlersErrors.CreateSnapshotError = false
	stepHandlersErrors.RemoveVolumeError = false
	stepHandlersErrors.VolumeInstancesError = false
	stepHandlersErrors.BadVolIDError = false
	stepHandlersErrors.NoCsiVolIDError = false
	stepHandlersErrors.WrongVolIDError = false
	stepHandlersErrors.WrongSystemError = false
	stepHandlersErrors.NoEndpointError = false
	stepHandlersErrors.NoUserError = false
	stepHandlersErrors.NoPasswordError = false
	stepHandlersErrors.NoSysNameError = false
	stepHandlersErrors.NoAdminError = false
	stepHandlersErrors.WrongSysNameError = false
	stepHandlersErrors.NoVolumeIDError = false
	stepHandlersErrors.SetVolumeSizeError = false
	stepHandlersErrors.systemNameMatchingError = false
	stepHandlersErrors.LegacyVolumeConflictError = false
	stepHandlersErrors.VolumeIDTooShortError = false
	stepHandlersErrors.EmptyEphemeralID = false
	stepHandlersErrors.IncorrectEphemeralID = false
	stepHandlersErrors.TooManyDashesVolIDError = false
	stepHandlersErrors.CorrectFormatBadCsiVolID = false
	stepHandlersErrors.EmptySysID = false
	stepHandlersErrors.VolIDListEmptyError = false
	stepHandlersErrors.CreateVGSNoNameError = false
	stepHandlersErrors.CreateVGSNameTooLongError = false
	stepHandlersErrors.CreateVGSLegacyVol = false
	stepHandlersErrors.CreateVGSAcrossTwoArrays = false
	stepHandlersErrors.CreateVGSBadTimeError = false
	stepHandlersErrors.CreateSplitVGSError = false
	stepHandlersErrors.BadVolIDJSON = false
	stepHandlersErrors.BadMountPathError = false
	stepHandlersErrors.NoMountPathError = false
	stepHandlersErrors.NoVolIDError = false
	stepHandlersErrors.NoVolIDSDCError = false
	stepHandlersErrors.NoVolError = false
	stepHandlersErrors.PeerMdmError = false
	stepHandlersErrors.CreateVolumeError = false
	stepHandlersErrors.BadRemoteSystemIDError = false
	stepHandlersErrors.NoProtectionDomainError = false
	stepHandlersErrors.GetReplicationConsistencyGroupsError = false
	stepHandlersErrors.ReplicationConsistencyGroupError = false
	stepHandlersErrors.ReplicationPairError = false
	stepHandlersErrors.GetReplicationPairError = false
	stepHandlersErrors.EmptyParametersListError = false
	stepHandlersErrors.RemoteReplicationConsistencyGroupError = false
	stepHandlersErrors.RemoteRCGBadNameError = false
	stepHandlersErrors.RemoveRCGError = false
	stepHandlersErrors.NoDeleteReplicationPair = false
	stepHandlersErrors.BadRemoteSystem = false
	stepHandlersErrors.ExecuteActionError = false
	stepHandlersErrors.StorageGroupAlreadyExists = false
	stepHandlersErrors.StorageGroupAlreadyExistsUnretriavable = false
	stepHandlersErrors.ReplicationGroupAlreadyDeleted = false
	stepHandlersErrors.ReplicationPairAlreadyExists = false
	stepHandlersErrors.ReplicationPairAlreadyExistsUnretrievable = false
	stepHandlersErrors.SnapshotCreationError = false
	stepHandlersErrors.GetRCGByIdError = false
	sdcMappings = sdcMappings[:0]
	sdcMappingsID = ""
	return handler
}

func getRouter() http.Handler {
	scaleioRouter := mux.NewRouter()
	scaleioRouter.HandleFunc("/api/instances/{from}::{id}/action/{action}", handleAction)
	scaleioRouter.HandleFunc("/api/instances/{from}::{id}/relationships/{to}", handleRelationships)
	scaleioRouter.HandleFunc("/api/types/Volume/instances/action/queryIdByKey", handleQueryVolumeIDByKey)
	scaleioRouter.HandleFunc("/api/instances/{type}::{id}", handleInstances)
	scaleioRouter.HandleFunc("/api/login", handleLogin)
	scaleioRouter.HandleFunc("/api/version", handleVersion)
	scaleioRouter.HandleFunc("/api/types/System/instances", handleSystemInstances)
	scaleioRouter.HandleFunc("/api/types/Volume/instances", handleVolumeInstances)
	scaleioRouter.HandleFunc("/api/types/PeerMdm/instances", handlePeerMdmInstances)
	scaleioRouter.HandleFunc("/api/types/StoragePool/instances", handleStoragePoolInstances)
	scaleioRouter.HandleFunc("/api/types/ReplicationConsistencyGroup/instances", handleReplicationConsistencyGroupInstances)
	scaleioRouter.HandleFunc("/api/types/ReplicationPair/instances", handleReplicationPairInstances)
	scaleioRouter.HandleFunc("{Volume}/relationship/Statistics", handleVolumeStatistics)
	scaleioRouter.HandleFunc("/api/Volume/relationship/Statistics", handleVolumeStatistics)
	scaleioRouter.HandleFunc("{SdcGUID}/relationships/Sdc", handleSystemSdc)
	return scaleioRouter
}

func addPreConfiguredVolume(id, name string) {
	volumeIDToName[id] = name
	volumeNameToID[name] = id
	volumeIDToSizeInKB[id] = defaultVolumeSize
	volumeIDToAncestorID[id] = ""
	volumeIDToConsistencyGroupID[id] = ""
	volumeIDToReplicationState[id] = unmarkedForReplication
}

func removePreConfiguredVolume(id string) {
	name := volumeIDToName[id]
	if name != "" {
		delete(volumeIDToName, id)
		delete(volumeNameToID, name)
	}
}

// handle implements GET /api/types/StoragePool/instances
func handleVolumeStatistics(w http.ResponseWriter, r *http.Request) {
	if stepHandlersErrors.PodmonVolumeStatisticsError {
		writeError(w, "induced error", http.StatusRequestTimeout, codes.Internal)
		return
	}
	returnJSONFile("features", "get_volume_statistics.json", w, nil)
}

func handleSystemSdc(w http.ResponseWriter, r *http.Request) {
	if stepHandlersErrors.GetSystemSdcError {
		writeError(w, "induced error", http.StatusRequestTimeout, codes.Internal)
		return
	}
	returnJSONFile("features", "get_sdc_instances.json", w, nil)
}

// handleLogin implements GET /api/login
func handleLogin(w http.ResponseWriter, r *http.Request) {
	u, p, ok := r.BasicAuth()
	if !ok || len(strings.TrimSpace(u)) < 1 || len(strings.TrimSpace(p)) < 1 {
		w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
		w.WriteHeader(http.StatusUnauthorized)
		returnJSONFile("features", "authorization_failure.json", w, nil)
		return
	}
	if testControllerHasNoConnection {
		w.WriteHeader(http.StatusRequestTimeout)
		return
	}
	w.Write([]byte("YWRtaW46MTU0MTU2MjIxOTI5MzpmODkxNDVhN2NkYzZkNGNkYjYxNGE0OGRkZGE3Zjk4MA"))
}

// handleLogin implements GET /api/version
func handleVersion(w http.ResponseWriter, r *http.Request) {
	if testControllerHasNoConnection {
		w.WriteHeader(http.StatusRequestTimeout)
		return
	}
	w.Write([]byte("2.5"))
}

// handleSystemInstances implements GET /api/types/System/instances
func handleSystemInstances(w http.ResponseWriter, r *http.Request) {
	if stepHandlersErrors.PodmonNodeProbeError {
		writeError(w, "PodmonNodeProbeError", http.StatusRequestTimeout, codes.Internal)
		return
	}
	if stepHandlersErrors.PodmonControllerProbeError {
		writeError(w, "PodmonControllerProbeError", http.StatusRequestTimeout, codes.Internal)
		return
	}
	if stepHandlersErrors.BadRemoteSystemIDError {
		returnJSONFile("features", "get_primary_system_instance.json", w, nil)
		return
	}
	if stepHandlersErrors.systemNameMatchingError {
		count++
	}
	if count == 2 || stepHandlersErrors.WrongSysNameError {
		fmt.Printf("DEBUG send bad system\n")
		returnJSONFile("features", "bad_system.json", w, nil)
		count = 0
	} else {
		returnJSONFile("features", "get_system_instances.json", w, nil)
	}
}

// handle PeerMDM instances implements GET /api/types/PeerMdm/instances
func handlePeerMdmInstances(w http.ResponseWriter, r *http.Request) {
	if stepHandlersErrors.PeerMdmError {
		writeError(w, "PeerMdmError", http.StatusRequestTimeout, codes.Internal)
		return
	}
	returnJSONFile("features", "get_peer_mdms.json", w, nil)
}

// handleStoragePoolInstances implements GET /api/types/StoragePool/instances
func handleStoragePoolInstances(w http.ResponseWriter, r *http.Request) {
	if stepHandlersErrors.GetStoragePoolsError {
		writeError(w, "induced error", http.StatusRequestTimeout, codes.Internal)
		return
	}
	returnJSONFile("features", "get_storage_pool_instances.json", w, nil)
}

func returnJSONFile(directory, filename string, w http.ResponseWriter, replacements map[string]string) (jsonBytes []byte) {
	jsonBytes, err := ioutil.ReadFile(filepath.Join(directory, filename))
	if err != nil {
		log.Printf("Couldn't read %s/%s\n", directory, filename)
		if w != nil {
			w.WriteHeader(http.StatusNotFound)
		}
		return make([]byte, 0)
	}
	if replacements != nil {
		jsonString := string(jsonBytes)
		for key, value := range replacements {
			jsonString = strings.Replace(jsonString, key, value, -1)
		}
		if debug {
			log.Printf("Edited payload:\n%s\n", jsonString)
		}
		jsonBytes = []byte(jsonString)
	}
	if debug {
		log.Printf("jsonBytes:\n%s\n", jsonBytes)
	}
	if w != nil {
		_, err = w.Write(jsonBytes)
		if err != nil {
			log.Printf("Couldn't write to ResponseWriter")
			w.WriteHeader(http.StatusInternalServerError)
			return make([]byte, 0)
		}
	}
	return jsonBytes
}

// Map of volume ID to name
var volumeIDToName map[string]string

// Map of volume name to ID
var volumeNameToID map[string]string

// Map of volume ID to ancestor ID
var volumeIDToAncestorID map[string]string

// Map of volume ID to consistency group ID
var volumeIDToConsistencyGroupID map[string]string

// Map of volume ID to size in KB
var volumeIDToSizeInKB map[string]string

// Map of volume ID to Replication State
var volumeIDToReplicationState map[string]string

// Map of Replication Consistency Group ID to name
var rcgIDToName map[string]string

// Map of Replication Consistency Group Name to ID
var rcgNameToID map[string]string

// Map of ReplicationPair ID to Name
var replicationPairIDToName map[string]string

// Map of ReplicatPair Name to ID
var replicationPairNameToID map[string]string

// Map of ReplicationPair ID to Source Volume
var replicationPairIDToSourceVolume map[string]string

// Map of ReplicationPair ID to Destination Volume
var replicationPairIDToDestinationVolume map[string]string

// Replication group mode to replace for.
var replicationGroupConsistMode string

// Replication group state to replace for.
var replicationGroupState string

var rcgIDtoDestinationVolumes map[string][]string

var replicationConsistencyGroups map[string]map[string]string

// handleVolumeInstances handles listing all volumes or creating a volume
func handleVolumeInstances(w http.ResponseWriter, r *http.Request) {
	if volumeIDToName == nil {
		volumeIDToName = make(map[string]string)
		volumeIDToAncestorID = make(map[string]string)
		volumeNameToID = make(map[string]string)
		volumeIDToConsistencyGroupID = make(map[string]string)
		volumeIDToSizeInKB = make(map[string]string)
		volumeIDToReplicationState = make(map[string]string)
		rcgIDToName = make(map[string]string)
		rcgNameToID = make(map[string]string)
		replicationConsistencyGroups = make(map[string]map[string]string)
		replicationPairIDToName = make(map[string]string)
		replicationPairNameToID = make(map[string]string)
		replicationPairIDToSourceVolume = make(map[string]string)
		replicationPairIDToDestinationVolume = make(map[string]string)
		rcgIDtoDestinationVolumes = make(map[string][]string)

	}
	if stepHandlersErrors.VolumeInstancesError {
		writeError(w, "induced error", http.StatusRequestTimeout, codes.Internal)
		return
	}

	if stepHandlersErrors.SnapshotCreationError {
		writeError(w, "RCG snapshot not created", http.StatusRequestTimeout, codes.Internal)
		return
	}

	switch r.Method {

	// Post is CreateVolume; here just return a volume id encoded from the name
	case http.MethodPost:
		if stepHandlersErrors.CreateVolumeError {
			writeError(w, "create volume induced error", http.StatusRequestTimeout, codes.Internal)
			return
		}
		req := types.VolumeParam{}
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			log.Printf("error decoding json: %s\n", err.Error())
		}
		fmt.Printf("POST to create volume name %s\n", req.Name)
		if volumeNameToID[req.Name] != "" {
			w.WriteHeader(http.StatusInternalServerError)
			// duplicate volume name response
			log.Printf("request for volume creation of duplicate name: %s\n", req.Name)
			resp := new(types.Error)
			resp.Message = sioGatewayVolumeNameInUse
			resp.HTTPStatusCode = http.StatusInternalServerError
			resp.ErrorCode = 6
			encoder := json.NewEncoder(w)
			err = encoder.Encode(resp)
			if err != nil {
				log.Printf("error encoding json: %s\n", err.Error())
			}
			return
		}
		// good response
		resp := new(types.VolumeResp)
		resp.ID = hex.EncodeToString([]byte(req.Name))
		fmt.Printf("Generated volume ID %s Name %s\n", resp.ID, req.Name)
		volumeIDToName[resp.ID] = req.Name
		volumeNameToID[req.Name] = resp.ID
		volumeIDToAncestorID[resp.ID] = "null"
		volumeIDToConsistencyGroupID[resp.ID] = "null"
		volumeIDToSizeInKB[resp.ID] = req.VolumeSizeInKb
		volumeIDToReplicationState[resp.ID] = unmarkedForReplication
		if debug {
			log.Printf("request name: %s id: %s\n", req.Name, resp.ID)
		}
		encoder := json.NewEncoder(w)
		err = encoder.Encode(resp)
		if err != nil {
			log.Printf("error encoding json: %s\n", err.Error())
		}

		log.Printf("end make volumes")
	// Read all the Volumes
	case http.MethodGet:
		instances := make([]*types.Volume, 0)
		for id, name := range volumeIDToName {
			name = id
			replacementMap := make(map[string]string)
			replacementMap["__ID__"] = id
			replacementMap["__NAME__"] = name
			replacementMap["__MAPPED_SDC_INFO__"] = getSdcMappings(id)
			replacementMap["__ANCESTOR_ID__"] = volumeIDToAncestorID[id]
			replacementMap["__CONSISTENCY_GROUP_ID__"] = volumeIDToConsistencyGroupID[id]
			replacementMap["__SIZE_IN_KB__"] = volumeIDToSizeInKB[id]
			replacementMap["__VOLUME_REPLICATION_STATE__"] = volumeIDToReplicationState[id]
			data := returnJSONFile("features", "volume.json.template", nil, replacementMap)
			vol := new(types.Volume)
			err := json.Unmarshal(data, vol)
			if err != nil {
				log.Printf("error unmarshalling json: %s\n", string(data))
			}
			instances = append(instances, vol)
		}
		encoder := json.NewEncoder(w)
		err := encoder.Encode(instances)
		if err != nil {
			log.Printf("error encoding json: %s\n", err)
		}
	}
}

const remoteRCGID = "d02aebc400000000"
const unmarkedForReplication = "UnmarkedForReplication"
const defaultVolumeSize = "33554432"

func handleReplicationConsistencyGroupInstances(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		if stepHandlersErrors.ReplicationConsistencyGroupError {
			writeError(w, "create rcg induced error", http.StatusRequestTimeout, codes.Internal)
			return
		}
		req := types.ReplicationConsistencyGroupCreatePayload{}
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			log.Printf("error decoding json: %s\n", err.Error())
		}
		fmt.Printf("POST to ReplicationConsistencyGroup %s\n", req.Name)
		if rcgNameToID[req.Name] != "" {
			w.WriteHeader(http.StatusInternalServerError)
			// duplicate rcg name response
			log.Printf("request for rcg creation of duplicate name: %s\n", req.Name)
			resp := new(types.Error)
			resp.Message = "The Replication Consistency Group already exists"
			resp.HTTPStatusCode = http.StatusInternalServerError
			resp.ErrorCode = 6
			encoder := json.NewEncoder(w)
			err = encoder.Encode(resp)
			if err != nil {
				log.Printf("error encoding json: %s\n", err.Error())
			}
			return
		}
		// good response
		resp := new(types.ReplicationConsistencyGroup)
		resp.ID = hex.EncodeToString([]byte(req.Name))
		fmt.Printf("Generated rcg ID %s Name %s\n", resp.ID, req.Name)
		rcgIDToName[resp.ID] = req.Name
		rcgNameToID[req.Name] = resp.ID
		// add in remote RCG, unless error
		rcgIDToName[remoteRCGID] = "rem-" + req.Name
		rcgNameToID["rem-"+req.Name] = req.Name

		replicationConsistencyGroups[resp.ID] = make(map[string]string)
		replicationConsistencyGroups[resp.ID]["Name"] = req.Name
		replicationConsistencyGroups[resp.ID]["ID"] = resp.ID
		replicationConsistencyGroups[resp.ID]["ProtectionGroup"] = req.ProtectionDomainId
		replicationConsistencyGroups[resp.ID]["RemoteProtectionGroup"] = req.RemoteProtectionDomainId

		if debug {
			log.Printf("request name: %s id: %s\n", req.Name, resp.ID)
		}

		if stepHandlersErrors.StorageGroupAlreadyExists || stepHandlersErrors.StorageGroupAlreadyExistsUnretriavable {
			writeError(w, "The Replication Consistency Group already exists", http.StatusRequestTimeout, codes.Internal)
			return
		}

		encoder := json.NewEncoder(w)
		err = encoder.Encode(resp)
		if err != nil {
			log.Printf("error encoding json: %s\n", err.Error())
		}
	case http.MethodGet:
		if stepHandlersErrors.GetReplicationConsistencyGroupsError {
			writeError(w, "could not GET ReplicationConsistencyGroups", http.StatusRequestTimeout, codes.Internal)
			return
		}
		instances := make([]*types.ReplicationConsistencyGroup, 0)
		for id, name := range rcgIDToName {
			if stepHandlersErrors.StorageGroupAlreadyExistsUnretriavable {
				continue
			}

			replacementMap := make(map[string]string)
			replacementMap["__ID__"] = id
			replacementMap["__NAME__"] = name
			if stepHandlersErrors.RemoteRCGBadNameError {
				replacementMap["__NAME__"] = "xxx"
			}
			replacementMap["__MODE__"] = replicationGroupConsistMode
			replacementMap["__PROTECTION_DOMAIN__"] = replicationConsistencyGroups[id]["ProtectionGroup"]
			replacementMap["__RM_PROTECTION_DOMAIN__"] = replicationConsistencyGroups[id]["RemoteProtectionGroup"]
			var data []byte
			if id == remoteRCGID {
				if stepHandlersErrors.RemoteReplicationConsistencyGroupError {
					writeError(w, "could not GET Remote ReplicationConsistencyGroup", http.StatusRequestTimeout, codes.Internal)
					return
				}
				data = returnJSONFile("features", "replication_consistency_group_reverse.template", nil, replacementMap)
			} else {
				data = returnJSONFile("features", "replication_consistency_group.template", nil, replacementMap)
			}
			fmt.Printf("RCG data %s\n", string(data))
			rcg := new(types.ReplicationConsistencyGroup)
			err := json.Unmarshal(data, rcg)
			if err != nil {
				log.Printf("error unmarshalling json: %s\n", string(data))
			}
			instances = append(instances, rcg)
		}
		encoder := json.NewEncoder(w)
		err := encoder.Encode(instances)
		if err != nil {
			log.Printf("error encoding json: %s\n", err)
		}

	}
}

func handleReplicationPairInstances(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		if stepHandlersErrors.ReplicationPairError {
			writeError(w, "POST ReplicationPair induced error", http.StatusRequestTimeout, codes.Internal)
			return
		}
		req := types.QueryReplicationPair{}
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			log.Printf("error decoding json: %s\n", err.Error())
		}
		fmt.Printf("POST to ReplicationPair %s Request %+v\n", req.Name, req)
		if replicationPairNameToID[req.Name] != "" {
			w.WriteHeader(http.StatusInternalServerError)
			// duplicate replication pair name response
			log.Printf("request for replication pair creation of duplicate name: %s\n", req.Name)
			resp := new(types.Error)
			resp.Message = "Replication Pair name already in use"
			resp.HTTPStatusCode = http.StatusInternalServerError
			resp.ErrorCode = 6
			encoder := json.NewEncoder(w)
			err = encoder.Encode(resp)
			if err != nil {
				log.Printf("error encoding json: %s\n", err.Error())
			}
			return
		}
		// good response
		resp := new(types.ReplicationPair)
		resp.ID = hex.EncodeToString([]byte(req.Name))
		fmt.Printf("Generated replicationPair ID %s Name %s Struct %+v\n", resp.ID, req.Name, req)
		replicationPairIDToName[resp.ID] = req.Name
		replicationPairIDToSourceVolume[resp.ID] = req.SourceVolumeID
		replicationPairIDToDestinationVolume[resp.ID] = req.DestinationVolumeID
		replicationPairNameToID[req.Name] = resp.ID

		volumeIDToReplicationState[req.SourceVolumeID] = "Replicated"
		volumeIDToReplicationState[req.DestinationVolumeID] = "Replicated"

		rcgIDtoDestinationVolumes[req.ReplicationConsistencyGroupID] = append(rcgIDtoDestinationVolumes[req.ReplicationConsistencyGroupID], req.DestinationVolumeID)

		if true {
			log.Printf("request name: %s id: %s sourceVolume %s\n", req.Name, resp.ID, req.SourceVolumeID)
		}

		if stepHandlersErrors.ReplicationPairAlreadyExists || stepHandlersErrors.ReplicationPairAlreadyExistsUnretrievable {
			writeError(w, "A Replication Pair for the specified local volume already exists", http.StatusRequestTimeout, codes.Internal)
			return
		}

		encoder := json.NewEncoder(w)
		err = encoder.Encode(resp)
		if err != nil {
			log.Printf("error encoding json: %s\n", err.Error())
		}
	case http.MethodGet:
		if stepHandlersErrors.GetReplicationPairError {
			writeError(w, "GET ReplicationPair induced error", http.StatusRequestTimeout, codes.Internal)
			return
		}
		instances := make([]*types.ReplicationPair, 0)
		for id, name := range replicationPairIDToName {
			if stepHandlersErrors.ReplicationPairAlreadyExistsUnretrievable {
				continue
			}

			replacementMap := make(map[string]string)
			replacementMap["__ID__"] = id
			replacementMap["__NAME__"] = name
			replacementMap["__SOURCE_VOLUME__"] = replicationPairIDToSourceVolume[id]
			replacementMap["__DESTINATION_VOLUME__"] = replicationPairIDToDestinationVolume[id]
			log.Printf("replicatPair replacementMap %v\n", replacementMap)
			data := returnJSONFile("features", "replication_pair.template", nil, replacementMap)
			log.Printf("replication-pair-data %s\n", string(data))
			pair := new(types.ReplicationPair)
			err := json.Unmarshal(data, pair)
			if err != nil {
				log.Printf("error unmarshalling json: %s\n", string(data))
			}
			log.Printf("replication-pair +%v", pair)
			instances = append(instances, pair)
		}
		encoder := json.NewEncoder(w)
		err := encoder.Encode(instances)
		if err != nil {
			log.Printf("error encoding json: %s\n", err)
		}
	}
}

func handleAction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	from := vars["from"]
	id := vars["id"]
	action := vars["action"]
	log.Printf("action from %s id %s action %s", from, id, action)
	switch action {
	case "addMappedSdc":
		if stepHandlersErrors.MapSdcError {
			writeError(w, "induced error", http.StatusRequestTimeout, codes.Internal)
			return
		}
		req := types.MapVolumeSdcParam{}
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			log.Printf("error decoding json: %s\n", err.Error())
		}
		fmt.Printf("SdcID: %s\n", req.SdcID)
		if req.SdcID == "d0f055a700000000" {
			sdcMappings = append(sdcMappings, types.MappedSdcInfo{SdcID: req.SdcID, SdcIP: "127.1.1.11"})
		}
		fmt.Printf("SdcID: %s\n", req.SdcID)
		if req.SdcID == "d0f055aa00000001" {
			sdcMappings = append(sdcMappings, types.MappedSdcInfo{SdcID: req.SdcID, SdcIP: "127.1.1.10"})
		}
	case "removeMappedSdc":
		if stepHandlersErrors.RemoveMappedSdcError {
			writeError(w, "induced error", http.StatusRequestTimeout, codes.Internal)
			return
		}
		req := types.UnmapVolumeSdcParam{}
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			log.Printf("error decoding json: %s\n", err.Error())
		}
		for i, val := range sdcMappings {
			if val.SdcID == req.SdcID {
				copy(sdcMappings[i:], sdcMappings[i+1:])
				sdcMappings = sdcMappings[:len(sdcMappings)-1]
			}
		}
	case "setMappedSdcLimits":
		if stepHandlersErrors.SDCLimitsError {
			writeError(w, "induced error", http.StatusRequestTimeout, codes.Internal)
			return
		}
		req := types.SetMappedSdcLimitsParam{}
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			log.Printf("error decoding json: %s\n", err.Error())
		}
		fmt.Printf("SdcID: %s\n", req.SdcID)
		if req.SdcID == "d0f055a700000000" {
			sdcMappings = append(sdcMappings, types.MappedSdcInfo{SdcID: req.SdcID})
		}
		fmt.Printf("BandwidthLimitInKbps: %s\n", req.BandwidthLimitInKbps)
		if req.BandwidthLimitInKbps == "10240" {
			sdcMappings = append(sdcMappings, types.MappedSdcInfo{SdcID: req.SdcID, LimitBwInMbps: 10})
		}
		fmt.Printf("IopsLimit: %s\n", req.IopsLimit)
		if req.IopsLimit == "11" {
			sdcMappings = append(sdcMappings, types.MappedSdcInfo{SdcID: req.SdcID, LimitIops: 11})
		}
	case "snapshotVolumes":
		if stepHandlersErrors.CreateSnapshotError {
			writeError(w, "induced error", http.StatusRequestTimeout, codes.Internal)
		}
		req := types.SnapshotVolumesParam{}
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			log.Printf("error decoding json: %s\n", err.Error())
		}
		for _, snapParam := range req.SnapshotDefs {
			// For now, only a single snapshot ID is supported

			id := snapParam.VolumeID

			cgValue := "f30216fb00000001"

			if snapParam.SnapshotName == "clone" || snapParam.SnapshotName == "volumeFromSnap" {
				id = "72cee42500000003"
			}
			if snapParam.SnapshotName == "invalid-clone" {
				writeError(w, "inducedError Volume not found", http.StatusRequestTimeout, codes.Internal)
				return
			}

			if stepHandlersErrors.WrongVolIDError {
				id = "72cee42500000002"
			}
			if stepHandlersErrors.FindVolumeIDError {
				id = "72cee42500000002"
				writeError(w, "inducedError Volume not found", http.StatusRequestTimeout, codes.Internal)
				return
			}

			// TWXXX EXPERIMENTAL
			volumeIDToName[id] = snapParam.SnapshotName
			volumeNameToID[snapParam.SnapshotName] = id
			volumeIDToAncestorID[id] = snapParam.VolumeID
			volumeIDToConsistencyGroupID[id] = cgValue
			volumeIDToSizeInKB[id] = defaultVolumeSize
			volumeIDToReplicationState[id] = unmarkedForReplication
		}

		if stepHandlersErrors.WrongVolIDError {
			returnJSONFile("features", "create_snapshot2.json", w, nil)
		}
		returnJSONFile("features", "create_snapshot.json", w, nil)
	case "removeVolume":
		if stepHandlersErrors.RemoveVolumeError {
			writeError(w, "inducedError", http.StatusRequestTimeout, codes.Internal)
		}
		name := volumeIDToName[id]
		volumeIDToName[id] = ""
		volumeIDToAncestorID[id] = ""
		volumeIDToConsistencyGroupID[id] = ""
		volumeIDToSizeInKB[id] = ""
		volumeIDToSizeInKB[id] = defaultVolumeSize
		volumeIDToReplicationState[id] = ""
		if name != "" {
			volumeNameToID[name] = ""
		}
	case "setVolumeSize":
		if stepHandlersErrors.SetVolumeSizeError {
			writeError(w, "induced error", http.StatusRequestTimeout, codes.Internal)
			return
		}
		req := types.SetVolumeSizeParam{}
		decoder := json.NewDecoder(r.Body)
		_ = decoder.Decode(&req)
		intValue, _ := strconv.Atoi(req.SizeInGB)
		volumeIDToSizeInKB[id] = strconv.Itoa(intValue / 1024)
	case "setVolumeName":
		//volumeIDToName[id] = snapParam.Name
		req := types.SetVolumeNameParam{}
		decoder := json.NewDecoder(r.Body)
		_ = decoder.Decode(&req)
		fmt.Printf("set volume name %s", req.NewName)
		volumeIDToName[id] = req.NewName
	case "removeReplicationConsistencyGroup":
		if stepHandlersErrors.RemoveRCGError {
			writeError(w, "inducedError", http.StatusRequestTimeout, codes.Internal)
			return
		}
		name := rcgIDToName[id]
		rcgIDToName[id] = ""
		if name != "" {
			rcgNameToID[name] = ""
		}
	case "removeReplicationPair":
		if stepHandlersErrors.NoDeleteReplicationPair {
			writeError(w, "pairs exist", http.StatusRequestTimeout, codes.Internal)
			return
		}
		sourceVolume := replicationPairIDToSourceVolume[id]
		destVolume := replicationPairIDToDestinationVolume[id]
		fmt.Printf("sourceVolume %s\n", sourceVolume)
		fmt.Printf("volumeIDToReplicationState %+v\n", volumeIDToReplicationState)
		volumeIDToReplicationState[sourceVolume] = unmarkedForReplication
		volumeIDToReplicationState[destVolume] = unmarkedForReplication
		name := replicationPairIDToName[id]
		delete(replicationPairIDToName, id)
		delete(replicationPairIDToSourceVolume, id)
		delete(replicationPairIDToDestinationVolume, id)
		delete(replicationPairNameToID, name)
	case "createReplicationConsistencyGroupSnapshots":
		if stepHandlersErrors.ExecuteActionError {
			writeError(w, "could not execute RCG action", http.StatusRequestTimeout, codes.Internal)
			return
		}
		// volumeIDToAncestorID[id] = "null"
		snapshotGroupID := uuid.New().String()
		resp := types.CreateReplicationConsistencyGroupSnapshotResp{}
		// snapshotID := hex.EncodeToString([]byte(snapshotName))
		resp.SnapshotGroupID = snapshotGroupID

		for key, val := range rcgIDtoDestinationVolumes {
			fmt.Printf("RCG ID %s, Vols %+v\n", key, val)

			for _, vol := range val {
				volName := uuid.New().String()
				volumeIDToName[volName] = volName
				volumeNameToID[volName] = volName
				volumeIDToAncestorID[volName] = vol
				volumeIDToConsistencyGroupID[volName] = snapshotGroupID
				volumeIDToSizeInKB[volName] = volumeIDToSizeInKB[vol]
				volumeIDToReplicationState[volName] = unmarkedForReplication

			}
		}

		encoder := json.NewEncoder(w)
		err := encoder.Encode(resp)
		if err != nil {
			log.Printf("error encoding json: %s\n", err)
		}
	case "switchoverReplicationConsistencyGroup":
		fallthrough
	case "failoverReplicationConsistencyGroup":
		fallthrough
	case "restoreReplicationConsistencyGroup":
		fallthrough
	case "reverseReplicationConsistencyGroup":
		fallthrough
	case "resumeReplicationConsistencyGroup":
		fallthrough
	case "pauseReplicationConsistencyGroup":
		if stepHandlersErrors.ExecuteActionError {
			writeError(w, "could not execute RCG action", http.StatusRequestTimeout, codes.Internal)
			return
		}
	}
}

func getSdcMappings(volumeID string) string {
	var bytes []byte
	var err error
	if sdcMappingsID == "" || volumeID == sdcMappingsID {
		bytes, err = json.Marshal(&sdcMappings)
	} else {
		var emptyMappings []types.MappedSdcInfo
		bytes, err = json.Marshal(&emptyMappings)
	}
	if err != nil {
		log.Printf("Json marshalling error: %s", err.Error())
		return ""
	}
	if debug {
		fmt.Printf("sdcMappings: %s\n", string(bytes))
	}
	return string(bytes)
}

func handleRelationships(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	from := vars["from"]
	id := vars["id"]
	to := vars["to"]
	log.Printf("relationship from %s id %s to %s", from, id, to)
	switch to {
	case "Sdc":
		if stepHandlersErrors.GetSdcInstancesError {
			writeError(w, "induced error", http.StatusRequestTimeout, codes.Internal)
		} else if stepHandlersErrors.PodmonFindSdcError {
			writeError(w, "PodmonFindSdcError", http.StatusRequestTimeout, codes.Internal)
		} else if stepHandlersErrors.PodmonNoSystemError {
			writeError(w, "PodmonNoSystemError", http.StatusRequestTimeout, codes.Internal)
		} else if stepHandlersErrors.PodmonControllerProbeError {
			writeError(w, "PodmonControllerProbeError", http.StatusRequestTimeout, codes.Internal)
			return
		} else if stepHandlersErrors.PodmonNodeProbeError {
			writeError(w, "PodmonNodeProbeError", http.StatusRequestTimeout, codes.Internal)
			return
		}
		returnJSONFile("features", "get_sdc_instances.json", w, nil)
	case "Statistics":
		if stepHandlersErrors.GetStatisticsError {
			writeError(w, "induced error", http.StatusRequestTimeout, codes.Internal)
			return
		}
		if from == "System" {
			returnJSONFile("features", "get_system_statistics.json", w, nil)
		} else if from == "StoragePool" {
			returnJSONFile("features", "get_storage_pool_statistics.json", w, nil)
		} else if from == "Volume" {
			if stepHandlersErrors.PodmonVolumeStatisticsError {
				writeError(w, "PodmonVolumeStatisticsError", http.StatusRequestTimeout, codes.Internal)
				return
			}
			returnJSONFile("features", "get_volume_statistics.json", w, nil)
		} else {
			writeError(w, "Unsupported relationship from type", http.StatusRequestTimeout, codes.Internal)
		}
	case "ProtectionDomain":
		if stepHandlersErrors.NoProtectionDomainError {
			writeError(w, "induced error NoProtectionDomainError", http.StatusRequestTimeout, codes.Internal)
			return
		}

		if from == "System" {
			returnJSONFile("features", "get_system_instances.json", w, nil)
		}

		returnJSONFile("features", "get_protection_domains.json", w, nil)
	case "ReplicationPair":
		if stepHandlersErrors.GetReplicationPairError {
			writeError(w, "GET ReplicationPair induced error", http.StatusRequestTimeout, codes.Internal)
			return
		}

		instances := make([]*types.ReplicationPair, 0)
		for id, name := range replicationPairIDToName {
			replacementMap := make(map[string]string)
			replacementMap["__ID__"] = id
			replacementMap["__NAME__"] = name
			replacementMap["__SOURCE_VOLUME__"] = replicationPairIDToSourceVolume[id]
			replacementMap["__DESTINATION_VOLUME__"] = replicationPairIDToDestinationVolume[id]
			data := returnJSONFile("features", "replication_pair.template", nil, replacementMap)
			pair := new(types.ReplicationPair)
			err := json.Unmarshal(data, pair)
			if err != nil {
				log.Printf("error unmarshalling json: %s\n", string(data))
			}
			log.Printf("pair +%v", pair)
			instances = append(instances, pair)
		}
		encoder := json.NewEncoder(w)
		err := encoder.Encode(instances)
		if err != nil {
			log.Printf("error encoding json: %s\n", err)
		}
	default:
		writeError(w, "Unsupported relationship to type", http.StatusRequestTimeout, codes.Internal)
	}
}

// handleInstances will retrieve specific instances
func handleInstances(w http.ResponseWriter, r *http.Request) {
	if stepHandlersErrors.BadVolIDError {
		writeError(w, "id must be a hexadecimal number", http.StatusRequestTimeout, codes.InvalidArgument)
		return
	}

	if stepHandlersErrors.GetVolByIDError {
		writeError(w, "induced error", http.StatusRequestTimeout, codes.Internal)
		return
	}
	if stepHandlersErrors.NoVolumeIDError {
		writeError(w, "volume ID is required", http.StatusRequestTimeout, codes.InvalidArgument)
		return
	}

	if stepHandlersErrors.SIOGatewayVolumeNotFoundError {
		writeError(w, "Could not find the volume", http.StatusRequestTimeout, codes.Internal)
		return
	}

	if stepHandlersErrors.ReplicationGroupAlreadyDeleted {
		writeError(w, "The Replication Consistency Group was not found", http.StatusRequestTimeout, codes.Internal)
		return
	}

	vars := mux.Vars(r)
	objType := vars["type"]
	id := vars["id"]
	id = extractIDFromStruct(id)
	log.Printf("handle instances type %s id %s\n", objType, id)
	switch objType {
	case "Volume":
		if id != "9999" {
			if volumeIDToName[id] == "" {
				// TWXXX EXPERIMENAL
				//fmt.Printf("volumeIDToName %v volumeNameToID %v\n", volumeIDToName, volumeNameToID)
				//writeError(w, "volume not found (no name): "+id, http.StatusNotFound, codes.NotFound)
				volumeIDToName[id] = "vol" + id
				volumeIDToSizeInKB[id] = defaultVolumeSize
				volumeIDToReplicationState[id] = unmarkedForReplication
			}
			log.Printf("Get id %s for %s\n", id, objType)
			replacementMap := make(map[string]string)
			replacementMap["__ID__"] = id
			replacementMap["__NAME__"] = volumeIDToName[id]
			replacementMap["__MAPPED_SDC_INFO__"] = getSdcMappings(id)
			replacementMap["__ANCESTOR_ID__"] = volumeIDToAncestorID[id]
			replacementMap["__CONSISTENCY_GROUP_ID__"] = volumeIDToConsistencyGroupID[id]
			replacementMap["__SIZE_IN_KB__"] = volumeIDToSizeInKB[id]
			replacementMap["__VOLUME_REPLICATION_STATE__"] = volumeIDToReplicationState[id]
			returnJSONFile("features", "volume.json.template", w, replacementMap)
		} else {
			log.Printf("Did not find id %s for %s\n", id, objType)
			writeError(w, "volume not found: "+id, http.StatusNotFound, codes.NotFound)
		}
	case "ReplicationConsistencyGroup":
		if stepHandlersErrors.GetRCGByIdError {
			writeError(w, "could not GET RCG by ID", http.StatusRequestTimeout, codes.Internal)
			return
		}

		replacementMap := make(map[string]string)
		replacementMap["__ID__"] = id
		replacementMap["__NAME__"] = rcgIDToName[id]
		replacementMap["__MODE__"] = replicationGroupConsistMode

		if replicationGroupState == "Normal" {
			replacementMap["__STATE__"] = "Ok"
		} else {
			replacementMap["__STATE__"] = "StoppedByUser"
			if replicationGroupState == "Failover" {
				replacementMap["__FO_TYPE__"] = "Failover"
				replacementMap["__FO_STATE__"] = "Done"
			} else if replicationGroupState == "Paused" {
				replacementMap["__P_MODE__"] = "Paused"
			}
		}

		returnJSONFile("features", "replication_consistency_group.template", w, replacementMap)
	}
}

// There are times when a struct {"id":"01234567890"} is sent for an id.
// This function extracts the id value
func extractIDFromStruct(id string) string {
	if !strings.HasPrefix(id, "{") {
		return id
	}
	// handle {"id":"012345678"} which seems to be passed in for this at times
	id = strings.Replace(id, "\"id\"", "", 1)
	id = strings.Replace(id, "{", "", 1)
	id = strings.Replace(id, "}", "", 1)
	id = strings.Replace(id, ":", "", 1)
	id = strings.Replace(id, "\"", "", -1)
	id = strings.Replace(id, "\n", "", -1)
	id = strings.Replace(id, " ", "", -1)
	return id
}

// Retrieve a volume by name
func handleQueryVolumeIDByKey(w http.ResponseWriter, r *http.Request) {
	if stepHandlersErrors.FindVolumeIDError {
		writeError(w, "induced error", http.StatusRequestTimeout, codes.Internal)
		return
	}
	req := new(types.VolumeQeryIDByKeyParam)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		log.Printf("error decoding json: %s\n", err.Error())
	}
	if volumeNameToID[req.Name] != "" {
		resp := new(types.VolumeResp)
		resp.ID = volumeNameToID[req.Name]
		log.Printf("found volume %s id %s\n", req.Name, volumeNameToID[req.Name])
		encoder := json.NewEncoder(w)
		if stepHandlersErrors.BadVolIDJSON {
			err = encoder.Encode("thisWill://causeUnmarshalErr")
		} else {
			err = encoder.Encode(resp.ID)
		}
		if err != nil {
			log.Printf("error encoding json: %s\n", err.Error())
		}
	} else {
		log.Printf("did not find volume %s\n", req.Name)
		volumeNameToID[req.Name] = ""
		writeError(w, fmt.Sprintf("Volume not found %s", req.Name), http.StatusNotFound, codes.NotFound)

	}
}

// Write an error code to the response writer
func writeError(w http.ResponseWriter, message string, httpStatus int, errorCode codes.Code) {
	w.WriteHeader(httpStatus)
	resp := new(types.Error)
	resp.Message = message
	resp.HTTPStatusCode = http.StatusNotFound
	resp.ErrorCode = int(errorCode)
	encoder := json.NewEncoder(w)
	err := encoder.Encode(resp)
	if err != nil {
		log.Printf("error encoding json: %s\n", err.Error())
	}
}
