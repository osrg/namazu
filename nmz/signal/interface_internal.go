package signal

import (
	"time"

	pb "github.com/osrg/namazu/nmz/util/pb"
)

// you don't have to take care of this interface, see Event and Action
type Signal interface {
	// RFC 4122 UUID
	//
	// json name: "uuid"
	ID() string

	// Entity ID string (e.g. "zksrv1")
	//
	// json name: "entity"
	EntityID() string

	// arrived time
	//
	// for event, only orchestrator should call this.
	// for action, only inspector should call this.
	ArrivedTime() time.Time

	// JSON map. PBEvent also implements this method, mainly for MongoDB storage.
	JSONMap() map[string]interface{}

	// debug string
	String() string
}

type ArrivalSignal interface {
	LoadJSONMap(map[string]interface{}) error
	SetArrivedTime(time.Time)
}

// Event for inspectors that *may* use ProtocolBuffers
type PBEvent interface {
	PBRequestMessage() *pb.InspectorMsgReq
}

// Action for inspectors that *may* use ProtocolBuffers
type PBAction interface {
	// can be nil, if Action.Event() does not implement PBEvent interface.
	PBResponseMessage() *pb.InspectorMsgRsp
}
