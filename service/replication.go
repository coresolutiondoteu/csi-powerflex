package service
import (
	"context"
	//"fmt"
	//"strings"

	//common "github.com/dell/dell-csi-extensions/common"
	//csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/dell/dell-csi-extensions/replication"
	//sio "github.com/dell/goscaleio"
	//siotypes "github.com/dell/goscaleio/types/v1"
        //"google.golang.org/grpc/codes"
        //"google.golang.org/grpc/status"
)

func (s *service) GetReplicationCapabilities(ctx context.Context, req *replication.GetReplicationCapabilityRequest) (*replication.GetReplicationCapabilityResponse, error) {
	Log.Printf("req GetReplicationCapabilities %+v", req)
	var rep = new(replication.GetReplicationCapabilityResponse)
	rep.Capabilities = []*replication.ReplicationCapability{
		{
			Type: &replication.ReplicationCapability_Rpc{
				Rpc: &replication.ReplicationCapability_RPC{
					Type: replication.ReplicationCapability_RPC_CREATE_REMOTE_VOLUME,
				},
			},
		},
		{
			Type: &replication.ReplicationCapability_Rpc{
				Rpc: &replication.ReplicationCapability_RPC{
					Type: replication.ReplicationCapability_RPC_CREATE_PROTECTION_GROUP,
				},
			},
		},
		{
			Type: &replication.ReplicationCapability_Rpc{
				Rpc: &replication.ReplicationCapability_RPC{
					Type: replication.ReplicationCapability_RPC_DELETE_PROTECTION_GROUP,
				},
			},
		},
		{
			Type: &replication.ReplicationCapability_Rpc{
				Rpc: &replication.ReplicationCapability_RPC{
					Type: replication.ReplicationCapability_RPC_REPLICATION_ACTION_EXECUTION,
				},
			},
		},
		{
			Type: &replication.ReplicationCapability_Rpc{
				Rpc: &replication.ReplicationCapability_RPC{
					Type: replication.ReplicationCapability_RPC_MONITOR_PROTECTION_GROUP,
				},
			},
		},
	}
	rep.Actions = []*replication.SupportedActions{
		{
			Actions: &replication.SupportedActions_Type{
				Type: replication.ActionTypes_FAILOVER_REMOTE,
			},
		},
		{
			Actions: &replication.SupportedActions_Type{
				Type: replication.ActionTypes_UNPLANNED_FAILOVER_LOCAL,
			},
		},
		{
			Actions: &replication.SupportedActions_Type{
				Type: replication.ActionTypes_REPROTECT_LOCAL,
			},
		},
		{
			Actions: &replication.SupportedActions_Type{
				Type: replication.ActionTypes_SUSPEND,
			},
		},
		{
			Actions: &replication.SupportedActions_Type{
				Type: replication.ActionTypes_RESUME,
			},
		},
		{
			Actions: &replication.SupportedActions_Type{
				Type: replication.ActionTypes_SYNC,
			},
		},
	}
	Log.Printf("rep GetReplicationCapabilities %+v", rep)
	return rep, nil
}

func (s *service) CreateStorageProtectionGroup(ctx context.Context, req *replication.CreateStorageProtectionGroupRequest) (*replication.CreateStorageProtectionGroupResponse, error) {
	Log.Printf("rep CreateStorageProtectionGroup %+v", req)
	return nil, nil
}

func (s *service) CreateRemoteVolume(ctx context.Context, req *replication.CreateRemoteVolumeRequest) (*replication.CreateRemoteVolumeResponse, error) {
	Log.Printf("rep CreateRemoteVolume %+v", req)
	return nil, nil
}

func (s *service) DeleteStorageProtectionGroup(ctx context.Context, req *replication.DeleteStorageProtectionGroupRequest) (*replication.DeleteStorageProtectionGroupResponse, error) {
	Log.Printf("rep DeleteStorageProtectionGroup %+v", req)
	return nil, nil
}

func (s *service) ExecuteAction(ctx context.Context, req *replication.ExecuteActionRequest) (*replication.ExecuteActionResponse, error) {
	Log.Printf("rep ExecuteAction %+v", req)
	return nil, nil
}

func (s *service) GetStorageProtectionGroupStatus(ctx context.Context, req *replication.GetStorageProtectionGroupStatusRequest) (*replication.GetStorageProtectionGroupStatusResponse, error) {
	Log.Printf("rep GetStorageProtectionGroupStatus %+v", req)
	return nil, nil
}






