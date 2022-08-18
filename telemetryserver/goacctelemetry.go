package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"syscall"
	"unsafe"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"golang.org/x/sys/windows"
)

const SPAGEFILEPHYSICS_STRUCT_SIZE uint32 = 712

type SPageFilePhysics struct {
	PacketId            int32
	Gas                 float32
	Brake               float32
	Fuel                float32
	Gear                int32
	Rpms                int32
	SteerAngle          float32
	SpeedKmh            float32
	Velocity            [3]float32
	AccG                [3]float32
	AheelSlip           [4]float32
	WheelLoad           [4]float32
	WheelsPressure      [4]float32
	WheelAngularSpeed   [4]float32
	TyreWear            [4]float32
	TyreDirtyLevel      [4]float32
	TyreCoreTemperature [4]float32
	CamberRAD           [4]float32
	SuspensionTravel    [4]float32
	Drs                 float32
	Tc                  float32
	Heading             float32
	Pitch               float32
	Roll                float32
	CgHeight            float32
	CarDamage           [5]float32
	NumberOfTyresOut    int32
	PitLimiterOn        int32
	Abs                 float32
	KersCharge          float32
	KersInput           float32
	AutoShifterOn       int32
	RideHeight          [2]float32
	TurboBoost          float32
	Ballast             float32
	AirDensity          float32
	AirTemp             float32
	RoadTemp            float32
	LocalAngularVel     [3]float32
	FinalFF             float32
	PerformanceMeter    float32

	EngineBrake      int32
	ErsRecoveryLevel int32
	ErsPowerLevel    int32
	ErsHeatCharging  int32
	ErsIsCharging    int32
	KersCurrentKJ    float32

	DrsAvailable int32
	DrsEnabled   int32

	BrakeTemp [4]float32
	Clutch    float32

	TyreTempI [4]float32
	TyreTempM [4]float32
	TyreTempO [4]float32

	IsAIControlled int32

	TyreContactPoint   [4][3]float32
	TyreContactNormal  [4][3]float32
	TyreContactHeading [4][3]float32

	BrakeBias float32

	LocalVelocity [3]float32

	P2PActivations int32
	P2PStatus      int32

	CurrentMaxRpm int32

	Mz        [4]float32
	Fx        [4]float32
	Fy        [4]float32
	SlipRatio [4]float32
	SlipAngle [4]float32

	TcInAction       int32
	AbsInAction      int32
	SuspensionDamage [4]float32
	TyreTemp         [4]float32
}

func readAccMemory(w http.ResponseWriter, r *http.Request) {

	physicsSharedMemName := "Local\\acpmf_physics"
	convertedName, err := windows.UTF16PtrFromString(physicsSharedMemName)

	memoryHandle, err := syscall.CreateFileMapping(syscall.Handle(windows.InvalidHandle), nil, syscall.PAGE_READWRITE, 0, SPAGEFILEPHYSICS_STRUCT_SIZE, convertedName)
	if err != nil {
		fmt.Println("Error when creating file mapping: " + err.Error())
	}

	ptrToPhysicsStruct, err := syscall.MapViewOfFile(memoryHandle, syscall.FILE_MAP_READ, 0, 0, uintptr(SPAGEFILEPHYSICS_STRUCT_SIZE))
	physicsStruct := (*SPageFilePhysics)(unsafe.Pointer(ptrToPhysicsStruct))

	restJSON(w, http.StatusOK, physicsStruct)
}

func restError(w http.ResponseWriter, code int, message string) {
	restJSON(w, code, map[string]string{"error": message})
}

func restJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func main() {
	fmt.Println("Assetto Corsa Competizione - Telemetry Server")
	var a SPageFilePhysics
	fmt.Println(unsafe.Sizeof(a))

	router := mux.NewRouter()
	router.HandleFunc("/api/physics", readAccMemory).Methods("GET")
	srv := &http.Server{
		Addr:    ":8080",
		Handler: handlers.CORS()(router),
	}
	log.Print("Listening on port 8080")
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {

		if pathTemplate, err := route.GetPathTemplate(); err == nil {
			log.Printf("http://localhost:8080%s %s", pathTemplate, route.GetName())
		}
		return nil
	})
	log.Fatal(srv.ListenAndServe())
}
