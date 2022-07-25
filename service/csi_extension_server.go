package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	podmon "github.com/dell/dell-csi-extensions/podmon"
	volumeGroupSnapshot "github.com/dell/dell-csi-extensions/volumeGroupSnapshot"
	sio "github.com/dell/goscaleio"
	siotypes "github.com/dell/goscaleio/types/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	//ExistingGroupID group id on powerflex array
	ExistingGroupID = "existingSnapshotGroupID"
)

func (s *service) ValidateVolumeHostConnectivity(ctx context.Context, req *podmon.ValidateVolumeHostConnectivityRequest) (*podmon.ValidateVolumeHostConnectivityResponse, error) {
	Log.Infof("ValidateVolumeHostConnectivity called %+v", req)
	rep := &podmon.ValidateVolumeHostConnectivityResponse{
		Messages: make([]string, 0),
	}

	if (len(req.GetVolumeIds()) == 0 || len(req.GetVolumeIds()) == 0) && req.GetNodeId() == "" {
		// This is a nop call just testing the interface is present
		rep.Messages = append(rep.Messages, "ValidateVolumeHostConnectivity is implemented")
		return rep, nil
	}

	// The NodeID for the VxFlex OS is the SdcGUID field.
	if req.GetNodeId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "The NodeID is a required field")
	}

	systemID := req.GetArrayId()
	if systemID == "" {
		if len(req.GetVolumeIds()) > 0 {
			systemID = s.getSystemIDFromCsiVolumeID(req.GetVolumeIds()[0])
		}
		if systemID == "" {
			systemID = s.opts.defaultSystemID
		}
	}

	// Do a probe of the requested system
	if err := s.requireProbe(ctx, systemID); err != nil {
		return nil, err
	}

	// First- check to see if the SDC is Connected or Disconnected.
	// Then retrieve the SDC and seet the connection state
	sdc, err := s.systems[systemID].FindSdc("SdcGUID", req.GetNodeId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "NodeID is invalid: %s - there is no corresponding SDC", req.GetNodeId())
	}
	connectionState := sdc.Sdc.MdmConnectionState
	rep.Messages = append(rep.Messages, fmt.Sprintf("SDC connection state: %s", connectionState))
	rep.Connected = (connectionState == "Connected")

	// Second- check to see if the Volumes have any I/O in the recent past.
	for _, volID := range req.GetVolumeIds() {
		// Probe system
		prevSystemID := systemID
		systemID = s.getSystemIDFromCsiVolumeID(volID)
		if systemID == "" {
			systemID = s.opts.defaultSystemID
		}
		if prevSystemID != systemID {
			if err := s.requireProbe(ctx, systemID); err != nil {
				rep.Messages = append(rep.Messages, fmt.Sprintf("Could not probe system: %s", volID))
				continue
			}
		}
		// Get the Volume
		vol, err := s.getVolByID(getVolumeIDFromCsiVolumeID(volID), systemID)
		if err != nil {
			rep.Messages = append(rep.Messages, fmt.Sprintf("Could not retrieve volume: %s", volID))
			continue
		}
		// Get the volume statistics
		volume := sio.NewVolume(s.adminClients[systemID])
		volume.Volume = vol
		stats, err := volume.GetVolumeStatistics()
		if err != nil {
			rep.Messages = append(rep.Messages, fmt.Sprintf("Could not retrieve volume statistics: %s", volID))
			continue
		}
		readCount := stats.UserDataReadBwc.NumOccured
		writeCount := stats.UserDataWriteBwc.NumOccured
		sampleSeconds := stats.UserDataWriteBwc.NumSeconds
		rep.Messages = append(rep.Messages, fmt.Sprintf("Volume %s writes %d reads %d for %d seconds",
			volID, writeCount, readCount, sampleSeconds))
		if (readCount + writeCount) > 0 {
			rep.IosInProgress = true
		}
	}

	Log.Infof("ValidateVolumeHostConnectivity reply %+v", rep)
	return rep, nil
}

func (s *service) CreateVolumeGroupSnapshot(ctx context.Context, req *volumeGroupSnapshot.CreateVolumeGroupSnapshotRequest) (*volumeGroupSnapshot.CreateVolumeGroupSnapshotResponse, error) {
	Log.Infof("CreateVolumeGroupSnapshot called with req: %v", req)

	err := validateCreateVGSreq(req)
	if err != nil {
		Log.Errorf("Error from CreateVolumeGroupSnapshot: %v ", err)
		return nil, err
	}

	//take first volume to calculate systemID. It is expected this systemID is consistent throughout
	systemID, err := s.getSystemID(req)
	if err != nil {
		Log.Errorf("Error from CreateVolumeGroupSnapshot: %v ", err)
		return nil, err
	}

	// Do a probe of the requested system
	if err := s.requireProbe(ctx, systemID); err != nil {
		return nil, err
	}

	Log.Infof("Creating Snapshot Consistency Group on system: %s", systemID)

	snapshotDefs, err := s.buildSnapshotDefs(req, systemID)

	if err != nil {
		Log.Errorf("Error from CreateVolumeGroupSnapshot: %v ", err)
		return nil, err
	}

	snapParam := &siotypes.SnapshotVolumesParam{SnapshotDefs: snapshotDefs}

	snapResponse, err := s.systems[systemID].CreateSnapshotConsistencyGroup(snapParam)
	if err != nil {
		var snapsThatFailed []string
		for _, snap := range snapshotDefs {
			snapsThatFailed = append(snapsThatFailed, snap.SnapshotName)
		}
		err = status.Errorf(codes.Internal, "Failed to create group with snapshots %s : %s", snapsThatFailed, err.Error())
		Log.Errorf("Error from CreateVolumeGroupSnapshot: %v ", err)
		return nil, err
	}
	Log.Infof("snapResponse is: %s", snapResponse)
	//populate response
	groupSnapshots, err := s.buildCreateVGSResponse(ctx, snapResponse, snapshotDefs, systemID)
	if err != nil {
		Log.Errorf("Error from CreateVolumeGroupSnapshot: %v ", err)
		return nil, err
	}

	//Check  Creation time, should be the same across all volumes
	err = checkCreationTime(groupSnapshots[0].CreationTime, groupSnapshots)
	if err != nil {
		return nil, err
	}

	resp := &volumeGroupSnapshot.CreateVolumeGroupSnapshotResponse{SnapshotGroupID: systemID + "-" + snapResponse.SnapshotGroupID, Snapshots: groupSnapshots, CreationTime: groupSnapshots[0].CreationTime}

	Log.Infof("CreateVolumeGroupSnapshot Response:  %#v", resp)
	return resp, nil
}

func checkCreationTime(time int64, snapshots []*volumeGroupSnapshot.Snapshot) error {
	Log.Infof("CheckCreationTime called with snapshots: %v", snapshots)
	for _, snap := range snapshots {
		if time != snap.CreationTime {
			err := status.Errorf(codes.Internal, "Creation time of snapshot %s, %d does not match with snapshot %s creation time %d. All snapshot creation times should be equal", snap.Name, snap.CreationTime, snapshots[0].Name, snapshots[0].CreationTime)
			Log.Errorf("Error from CheckCreationTime: %v ", err)
			return err
		}
		Log.Infof("CheckCreationTime: Creation time of %s is %d", snap.Name, time)

	}
	return nil
}

func (s *service) getSystemID(req *volumeGroupSnapshot.CreateVolumeGroupSnapshotRequest) (string, error) {
	//take first volume to calculate systemID. It is expected this systemID is consistent throughout
	systemID := s.getSystemIDFromCsiVolumeID(req.SourceVolumeIDs[0])
	if systemID == "" {
		// use default system
		systemID = s.opts.defaultSystemID
	}

	if systemID == "" {
		err := status.Error(codes.InvalidArgument, "systemID is not found in vol ID and there is no default system")
		Log.Errorf("Error from getSystemID: %v ", err)
		return systemID, err

	}

	return systemID, nil

}

//validate if request has source volumes, a VGS name, and VGS name length < 27 chars
func validateCreateVGSreq(req *volumeGroupSnapshot.CreateVolumeGroupSnapshotRequest) error {
	if len(req.SourceVolumeIDs) == 0 {
		err := status.Errorf(codes.InvalidArgument, "SourceVolumeIDs cannot be empty")
		Log.Errorf("Error from validateCreateVGSreq: %v ", err)
		return err
	}

	if req.Name == "" {
		err := status.Error(codes.InvalidArgument, "CreateVolumeGroupSnapshotRequest Name is not  set")
		Log.Warnf("Warning from validateCreateVGSreq: %v ", err)
	}

	//name must be less than 28 chars, because we name snapshots with -<index>, and index can at most be 3 chars
	if len(req.Name) > 27 {
		err := status.Errorf(codes.InvalidArgument, "Requested name %s longer than 27 character max", req.Name)
		Log.Errorf("Error from validateCreateVGSreq: %v ", err)
		return err
	}

	return nil
}

func (s *service) buildSnapshotDefs(req *volumeGroupSnapshot.CreateVolumeGroupSnapshotRequest, systemID string) ([]*siotypes.SnapshotDef, error) {

	snapshotDefs := make([]*siotypes.SnapshotDef, 0)

	for _, id := range req.SourceVolumeIDs {
		snapSystemID := strings.TrimSpace(s.getSystemIDFromCsiVolumeID(id))
		if snapSystemID != "" && snapSystemID != systemID {
			err := status.Errorf(codes.Internal, "Source volumes for volume group snapshot should be on the same system but vol %s is not on system: %s", id, systemID)
			Log.Errorf("Error from buildSnapshotDefs: %v \n", err)
			return nil, err
		}

		//legacy vol check
		err := s.checkVolumesMap(id)
		if err != nil {
			err = status.Errorf(codes.Internal, "checkVolumesMap for id: %s failed : %s", id, err.Error())
			Log.Errorf("Error from buildSnapshotDefs: %v ", err)
			return nil, err
		}

		volID := getVolumeIDFromCsiVolumeID(id)

		_, err = s.getVolByID(volID, systemID)
		if err != nil {
			err = status.Errorf(codes.Internal, "failure checking source volume status: %s", err.Error())
			Log.Errorf("Error from buildSnapshotDefs: %v ", err)
			return nil, err
		}

		snapDef := siotypes.SnapshotDef{VolumeID: volID, SnapshotName: ""}
		snapshotDefs = append(snapshotDefs, &snapDef)
	}

	return snapshotDefs, nil

}

//build the response for CreateVGS to return
func (s *service) buildCreateVGSResponse(ctx context.Context, snapResponse *siotypes.SnapshotVolumesResp, snapshotDefs []*siotypes.SnapshotDef, systemID string) ([]*volumeGroupSnapshot.Snapshot, error) {
	var groupSnapshots []*volumeGroupSnapshot.Snapshot
	for index, id := range snapResponse.VolumeIDList {
		idToQuery := systemID + "-" + id
		req := &csi.ListSnapshotsRequest{SnapshotId: idToQuery}
		lResponse, err := s.ListSnapshots(ctx, req)
		if err != nil {
			err = status.Errorf(codes.Internal, "Failed to get snapshot: %s", err.Error())
			Log.Errorf("Error from buildCreateVGSResponse: %v ", err)
			return nil, err
		}
		var arraySnapName string
		// ancestorvolumeid
		existingSnap, _ := s.adminClients[systemID].GetVolume("", id, lResponse.Entries[0].Snapshot.SourceVolumeId, "", true)
		for _, e := range existingSnap {
			if e.ID == id && e.ConsistencyGroupID == snapResponse.SnapshotGroupID {
				if e.Name == "" {
					Log.Infof("debug set snap name for [%s]", e.ID)
					arraySnapName = e.ID + "-snap-" + strconv.Itoa(index)
					tgtVol := sio.NewVolume(s.adminClients[systemID])
					tgtVol.Volume = e
					err := tgtVol.SetVolumeName(arraySnapName)
					if err != nil {
						Log.Errorf("Error setting name of snapshot id=%s name=%s %s", e.ID, arraySnapName, err.Error())
					}
				} else {
					Log.Infof("debug found snap name %s for %s", e.Name, e.ID)
					arraySnapName = e.Name
				}
			}
		}

		Log.Infof("Snapshot Name created for: %s is %s", lResponse.Entries[0].Snapshot.SnapshotId, arraySnapName)
		//need to convert time from seconds and nanoseconds to int64 nano seconds
		creationTime := lResponse.Entries[0].Snapshot.CreationTime.GetSeconds()*1000000000 + int64(lResponse.Entries[0].Snapshot.CreationTime.GetNanos())
		Log.Infof("Creation time is: %d\n", creationTime)
		snap := volumeGroupSnapshot.Snapshot{
			Name:          arraySnapName,
			CapacityBytes: lResponse.Entries[0].Snapshot.SizeBytes,
			SnapId:        lResponse.Entries[0].Snapshot.SnapshotId,
			SourceId:      systemID + "-" + lResponse.Entries[0].Snapshot.SourceVolumeId,
			ReadyToUse:    lResponse.Entries[0].Snapshot.ReadyToUse,
			CreationTime:  creationTime,
		}
		groupSnapshots = append(groupSnapshots, &snap)
	}

	return groupSnapshots, nil

}
