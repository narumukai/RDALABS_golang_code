package cleanup

import (
	"github.com/nautiluslabsco/null"
	"math"
	"time"

	"github.com/nautiluslabsco/ln/features/calc"
	"github.com/nautiluslabsco/ln/features/calc/calcapi"
	"github.com/nautiluslabsco/ln/shared/constants/labels"
	"github.com/nautiluslabsco/ln/shared/constants/units"
	"github.com/nautiluslabsco/ln/shared/models"
)

var cleanupFuncs = []CleanupFunc{
	{
		Comment:  "Filter period of weird shaft power / shaft speed NAUT-1439",
		Issue:    "NAUT-1439",
		ShipID:   calc.EagleJay,
		Start:    parseTime("2017-09-26 21:00"),
		End:      parseTime("2017-10-05 01:00"),
		Stage:    PreVesselAnatomyStage,
		CalcFunc: calc.SetNull(labels.ShaftSpeed, labels.ShaftPower),
	},
	{
		Comment:  "Remove large region of erroneous values NAUT-1434",
		Issue:    "NAUT-1434",
		ShipID:   7,
		Start:    parseTime("2018-09-10 01:00"),
		End:      parseTime("2018-10-31 01:00"),
		Stage:    PreVesselAnatomyStage,
		CalcFunc: calc.SetNull(labels.ShaftSpeed, labels.ShaftPower),
	},
	{
		Comment: "Scale elevated shaft power figures NAUT-1440",
		Issue:   "NAUT-1440",
		ShipID:  8,
		Start:   parseTime("2017-12-09 08:00"),
		End:     parseTime("2018-01-24 23:00"),
		Stage:   PreVesselAnatomyStage,
		CalcFunc: func(pc calcapi.PropertyCalc) {
			pc.SetProperty(labels.ShaftPower, pc.GetProperty(labels.ShaftPower)/2.84)
		},
	},
	{
		Comment:  "Large region of extremely elevated STW",
		Issue:    "NAUT-1523",
		ShipID:   16,
		Start:    parseTime("2016-12-21 12:00"),
		End:      parseTime("2017-01-01 06:00"),
		Stage:    PreVesselAnatomyStage,
		CalcFunc: calc.SetNull(labels.SpeedThroughWater),
	},
	{
		Issue:         "NAUT-2022",
		ShipID:        72,
		Start:         parseTime("2018-10-11 00:00"),
		End:           parseTime("2018-11-08 00:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc: calc.SetNull(
			"ME_HSFO_t_h",
			"ME_LSFO_t_h",
			"ME_MDO_t_h",
			"ME_MGO_t_h",
			"AE_HSFO_t_h",
			"AE_LSFO_t_h",
			"AE_MDO_t_h",
			"AE_MGO_t_h",
			"BLR_HSFO_t_h",
			"BLR_LSFO_t_h",
			"BLR_MDO_t_h",
			"BLR_MGO_t_h"),
	},
	{
		Issue:         "NAUT-2022",
		ShipID:        72,
		Start:         time.Time{},
		End:           time.Time{},
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc: func(pc calcapi.PropertyCalc) {
			if pc.GetProperty(labels.ShaftPower) > 37000 || pc.GetProperty(labels.ShaftSpeed) > 140 {
				pc.SetNullableProperty(labels.ShaftPower, models.NullValue())
				pc.SetNullableProperty(labels.ShaftSpeed, models.NullValue())
			}

			for _, lowLevelFlow := range []string{
				"ME_HSFO_t_h",
				"ME_LSFO_t_h",
				"ME_MDO_t_h",
				"ME_MGO_t_h",
				"AE_HSFO_t_h",
				"AE_LSFO_t_h",
				"AE_MDO_t_h",
				"AE_MGO_t_h",
				"BLR_HSFO_t_h",
				"BLR_LSFO_t_h",
				"BLR_MDO_t_h",
				"BLR_MGO_t_h",
			} {
				if pc.GetProperty(lowLevelFlow) > 2 && pc.GetProperty(labels.ShaftPower) < 3500 {
					pc.SetNullableProperty(lowLevelFlow, models.NullValue())
				}
				if pc.GetProperty(lowLevelFlow) < 0 || pc.GetProperty(lowLevelFlow) > 3 {
					pc.SetNullableProperty(lowLevelFlow, models.NullValue())
				}
			}
		},
	},
	{
		Issue:         "NAUT-1860",
		ShipID:        1,
		Start:         time.Time{},
		End:           time.Time{},
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc: func(pc calcapi.PropertyCalc) {
			pc.SetProperty(labels.SensorSTW, pc.GetProperty(labels.SpeedThroughWater))
			modeledSTW := copernicusModeledSTW(pc)
			if modeledSTW.Valid {
				pc.SetProperty(labels.ModeledSTW, modeledSTW.Float64)
			}
		},
	},
	{
		Issue:         "NAUT-1860",
		ShipID:        1,
		Start:         parseTime("2018-03-01 00:00"),
		End:           time.Time{},
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc: func(pc calcapi.PropertyCalc) {
			pc.SetProperty(labels.SpeedThroughWater, pc.GetProperty(labels.ModeledSTW))
		},
	},
	{
		Issue:         "NAUT-2289",
		ShipID:        36,
		Start:         parseTime("2019-08-10 00:00"),
		End:           parseTime("2019-08-10 02:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      RemoveBadGPS,
	},
	{
		Issue:         "NAUT-2472",
		ShipID:        18,
		Start:         parseTime("2019-02-06 22:00"),
		End:           parseTime("2019-02-07 03:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      RemoveBadGPS,
	},
	{
		Issue:         "NAUT-2259",
		ShipID:        45,
		Start:         parseTime("2019-07-27 21:00"),
		End:           time.Time{},
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc: func(pc calcapi.PropertyCalc) {
			power, _ := units.Convert(pc.GetProperty("PROPULSION SHAFT POWER"), units.MegaWatts.Id, units.KiloWatts.Id)
			power = power * 0.99
			pc.SetProperty(labels.ShaftPower, power)
		},
	},
	{
		Comment:       "Remove erroneous shaft values before cutoff date",
		Issue:         "DPI-723",
		ShipID:        calc.EpsMountHermon,
		Start:         parseTime("2020-01-01 00:00"),
		End:           parseTime("2020-07-02 00:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.SetNull(labels.ShaftPower, labels.ShaftSpeed),
	},
	{
		Comment:       "Remove erroneous shaft values before cutoff date",
		Issue:         "DPI-723",
		ShipID:        calc.EpsSolomonSea,
		Start:         parseTime("2020-01-01 00:00"),
		End:           parseTime("2020-05-25 00:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.SetNull(labels.ShaftPower, labels.ShaftSpeed),
	},
	{
		Comment:       "Remove erroneous latitude/longitude values",
		Issue:         "DPI-415",
		ShipID:        555128,
		Start:         parseTime("2020-02-29 12:50"),
		End:           parseTime("2020-02-29 13:10"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      RemoveBadGPS,
	},
	{
		Comment:       "Remove erroneous latitude/longitude values",
		Issue:         "DPI-809",
		ShipID:        59,
		Start:         parseTime("2020-08-22 10:50"),
		End:           parseTime("2020-08-22 11:10"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      RemoveBadGPS,
	},
	{
		Comment:       "Remove erroneous latitude/longitude values",
		Issue:         "DPI-734",
		ShipID:        376230,
		Start:         parseTime("2020-06-08 22:00"),
		End:           parseTime("2020-07-14 17:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      RemoveBadGPS,
	},
	{
		Comment:       "SOG sensor isn't working, so map over it with ObservedSpeed",
		Issue:         "DPI-807",
		ShipID:        7,
		Start:         parseTime("2020-07-22 14:00"),
		End:           parseTime("2020-09-22 00:00"), // see ENG-306
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc: func(pc calcapi.PropertyCalc) {
			if obs := pc.GetNullableProperty(labels.ObservedSpeed); obs.Present() {
				pc.SetPropertyWithUnit(labels.SpeedOverGround,
					obs.Value(), pc.GetUnitForProperty(labels.ObservedSpeed))
			}
		},
	},
	{
		Comment:       "Sensor data doesn't provide sign for position, but noons do",
		Issue:         "DPI-835, ENG-553",
		ShipID:        616,
		Start:         parseTime("2020-01-01 00:00"),
		End:           parseTime("2020-12-20 00:00"), // sign was fixed from here on
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc:      OverrideLatLonSign,
	},
	{
		Comment:       "Sensor data doesn't provide sign for position, but noons do",
		Issue:         "DPI-835, DPI-962",
		ShipID:        146207,
		Start:         parseTime("2020-01-01 00:00"),
		End:           parseTime("2020-10-20 00:00"), // when received first negative latitude value
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc:      OverrideLatLonSign,
	},
	{
		Comment:       "Generator Power tags were changed",
		Issue:         "DPI-922",
		ShipID:        207,
		Start:         parseTime("2020-01-01 00:00"),
		End:           parseTime("2020-09-30 06:00"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc:      OverrideChevronGeneratorPower,
	},
	{
		Comment:       "Generator Power tags were changed",
		Issue:         "DPI-922",
		ShipID:        896,
		Start:         parseTime("2020-01-01 00:00"),
		End:           parseTime("2020-09-13 16:00"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc:      OverrideChevronGeneratorPower,
	},
	{
		Comment:       "Generator Power tags were changed",
		Issue:         "DPI-922",
		ShipID:        971,
		Start:         parseTime("2020-01-01 00:00"),
		End:           parseTime("2020-09-27 00:00"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc:      OverrideChevronGeneratorPower,
	},
	{
		Comment:       "Position Sign tags not provided yet",
		Issue:         "DPI-920",
		ShipID:        207,
		Start:         parseTime("2020-01-01 00:00"),
		End:           parseTime("2021-01-01 00:00"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc:      OverrideLatLonSign,
	},
	{
		Comment:       "Position Sign tags not provided yet",
		Issue:         "DPI-920",
		ShipID:        896,
		Start:         parseTime("2020-01-01 00:00"),
		End:           parseTime("2021-01-01 00:00"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc:      OverrideLatLonSign,
	},
	{
		Comment:       "Position Sign tags not provided yet",
		Issue:         "DPI-920",
		ShipID:        971,
		Start:         parseTime("2020-01-01 00:00"),
		End:           parseTime("2021-01-01 00:00"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc:      OverrideLatLonSign,
	},
	{
		Comment:       "Correcting sign which is causing interpolation error",
		Issue:         "DPI-925",
		ShipID:        971,
		Start:         parseTime("2020-10-10 23:00"),
		End:           parseTime("2020-10-11 01:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      NegateLatitude,
	},
	{
		Comment:       "Correcting sign which is causing interpolation error",
		Issue:         "DPI-925",
		ShipID:        119,
		Start:         parseTime("2020-10-10 23:00"),
		End:           parseTime("2020-10-11 01:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      NegateLatitude,
		
	},
	
	
	{
		Comment:       "Correcting sign which is causing interpolation error",
		Issue:         "DPI-925",
		ShipID:        207,
		Start:         parseTime("2020-09-30 05:00"),
		End:           parseTime("2020-09-30 07:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      NegateLatitude,
	},
	{
		Comment:       "Correcting sign which is causing interpolation error",
		Issue:         "DPI-925",
		ShipID:        207,
		Start:         parseTime("2020-10-08 11:00"),
		End:           parseTime("2020-10-09 04:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      NegateLatitude,
	},
	{
		Comment: "Correcting sign after noon correction above " +
			"due to transition from N to S in the from Noon to Noon",
		Issue:         "DPI-925",
		ShipID:        896,
		Start:         parseTime("2020-09-18 05:00"),
		End:           parseTime("2020-09-18 18:00"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc:      NegateLatitude,
	},
	{
		Comment: "Correcting sign after noon correction above " +
			"due to transition from N to S in the from Noon to Noon",
		Issue:         "DPI-925",
		ShipID:        896,
		Start:         parseTime("2020-10-02 04:00"),
		End:           parseTime("2020-10-02 20:00"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc:      NegateLatitude,
	},
	{
		Comment:       "Generator Fuel Flow is in MT/hr prior to cutoff",
		Issue:         "ENG-397",
		ShipID:        calc.EpsPacificBeryl,
		Start:         parseTime("2020-01-01 00:00"),
		End:           parseTime("2020-10-08 00:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.EpsEnamorFixUnit("AE_HSFO_t_h", "AE_LSFO_t_h", "AE_MDO_t_h", "AE_MGO_t_h"),
	},
	{
		Comment:       "Jacaranda didn't always have a Fuel Outlet tag. So we 'fake' it to ensure fuel calculations occur",
		Issue:         "ENG-377",
		ShipID:        calc.Jacaranda,
		Start:         parseTime("2020-07-01 00:00"), // roughly when we started getting sensor data
		End:           parseTime("2020-11-05 00:00"), // autologger updated to use v2 samelectronics modbus
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.LoginPsuedoOutlet,
	},
	{
		Comment:       "Remove ESM erroneous fuel flow data before valid data is received",
		Issue:         "DPI-938",
		ShipID:        263, // Roberto
		Start:         parseTime("2020-08-25 00:00"),
		End:           parseTime("2020-10-21 00:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.SetNull("STBD VBU unit", "PORT VBU unit"),
	},
	{
		Comment:       "Remove ESM erroneous fuel flow data before valid data is received",
		Issue:         "DPI-938",
		ShipID:        196, // Red Marauder
		Start:         parseTime("2020-08-25 00:00"),
		End:           parseTime("2020-10-27 00:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.SetNull("STBD VBU unit", "PORT VBU unit"),
	},
	{
		Comment:       "Remove ESM erroneous fuel flow data before valid data is received",
		Issue:         "DPI-938",
		ShipID:        316, // Reference Point
		Start:         parseTime("2020-08-25 00:00"),
		End:           parseTime("2020-10-30 00:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.SetNull("STBD VBU unit", "PORT VBU unit"),
	},
	{
		Comment:       "Remove ESM erroneous fuel flow data before valid data is received",
		Issue:         "DPI-938",
		ShipID:        167, // Red Rum
		Start:         parseTime("2020-08-25 00:00"),
		End:           parseTime("2020-11-03 00:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.SetNull("STBD VBU unit", "PORT VBU unit"),
	},
	{
		Comment:       "CMA CGM Tenere - Remove erroneous draft values",
		Issue:         "DPI-960",
		ShipID:        181,
		Start:         parseTime("2020-09-17 07:00"),
		End:           parseTime("2020-09-18 01:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.SetNull(labels.DraftAft, labels.DraftFwd, labels.DraftMid1, labels.DraftMid2),
	},
	{
		Comment:       "Remove Hunter position data when invalid",
		Issue:         "ENG-383",
		ShipID:        calc.HunterIdun, // Idun
		Start:         parseTime("2020-08-23"),
		End:           parseTime("2020-10-22"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      RemoveBadGPS,
	},
	{
		Comment:       "Remove Hunter position data when invalid",
		Issue:         "ENG-383",
		ShipID:        calc.HunterFrigg, // Frigg
		Start:         parseTime("2020-08-23"),
		End:           parseTime("2020-10-22"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      RemoveBadGPS,
	},
	{
		Comment:       "Remove Hunter position data when invalid",
		Issue:         "ENG-383",
		ShipID:        calc.HunterFreya, // Freya
		Start:         parseTime("2020-08-23"),
		End:           parseTime("2020-10-22"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      RemoveBadGPS,
	},
	{
		Comment:       "Remove Hunter position data when invalid",
		Issue:         "ENG-383",
		ShipID:        calc.HunterFreya, // Freya
		Start:         parseTime("2020-12-10 00:00:00"),
		End:           parseTime("2020-12-17 17:00:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.SetNull("Voyage Location latitude", "Voyage Location longitude"),
	},
	{
		Comment:       "Remove Hunter position data when invalid",
		Issue:         "ENG-383",
		ShipID:        calc.HunterFreya, // Freya
		Start:         parseTime("2020-11-30 04:00:00"),
		End:           parseTime("2020-12-09 11:00:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      RemoveBadGPS,
	},
	{
		Comment:       "Chevron Asia Energy - Remove erroneous positions",
		Issue:         "ENG-449",
		ShipID:        207,
		Start:         parseTime("2020-10-15 00:00"),
		End:           parseTime("2020-11-12 02:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      RemoveBadGPS,
	},
	{
		Comment:       "Chevron Asia Vision - Remove erroneous positions",
		Issue:         "ENG-449",
		ShipID:        896,
		Start:         parseTime("2020-10-23 00:00"),
		End:           parseTime("2020-11-08 00:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      RemoveBadGPS,
	},
	{
		Comment:       "Chevron Asia Vision - Remove erroneous positions",
		Issue:         "ENG-449",
		ShipID:        896,
		Start:         parseTime("2020-11-11 04:00"),
		End:           parseTime("2020-11-15 02:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      RemoveBadGPS,
	},
	{
		Comment:       "Diamond Bulk Sincere Pisces - Alias alternative mode switch tags",
		Issue:         "ENG-476",
		ShipID:        calc.DbcSincerePisces,
		Start:         time.Time{}, // Unbounded
		End:           parseTime("2020-11-25 00:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc: func(pc calcapi.PropertyCalc) {
			// Sets the shared Hoppe mode switch tags from the DBC exclusive mode switch tags
			pc.SetNullableProperty("Main Engine on ULSFO", pc.GetNullableProperty("Main Engine using ULSFO"))
			pc.SetNullableProperty("Main Engine on MGO", pc.GetNullableProperty("Main Engine using MGO"))
			pc.SetNullableProperty("Main Engine on MDO", pc.GetNullableProperty("Main Engine using MDO"))
			pc.SetNullableProperty("Main Engine on LSMGO", pc.GetNullableProperty("Main Engine using LSMGO"))
			pc.SetNullableProperty("Main Engine on LSHFO", pc.GetNullableProperty("Main Engine using LSHFO"))
			pc.SetNullableProperty("Main Engine on HFO", pc.GetNullableProperty("Main Engine using HFO"))
		},
	},
	{
		Comment:       "Hunter fallback to voyage location gps must run before weather service",
		Issue:         "ENG-477",
		ShipID:        calc.HunterAtla,
		Start:         time.Time{}, // unbounded
		End:           time.Time{},
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.EnableFallbackToVoyageLocation,
	},
	{
		Comment:       "Hunter fallback to voyage location gps must run before weather service",
		Issue:         "ENG-477",
		ShipID:        calc.HunterDisen,
		Start:         time.Time{}, // unbounded
		End:           time.Time{},
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.EnableFallbackToVoyageLocation,
	},
	{
		Comment:       "Hunter fallback to voyage location gps must run before weather service",
		Issue:         "ENG-477",
		ShipID:        calc.HunterFreya,
		Start:         time.Time{}, // unbounded
		End:           parseTime("2020-10-27 11:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.EnableFallbackToVoyageLocation,
	},
	{
		Comment:       "Hunter fallback to voyage location gps must run before weather service",
		Issue:         "ENG-477",
		ShipID:        calc.HunterFrigg,
		Start:         time.Time{}, // unbounded
		End:           time.Time{},
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.EnableFallbackToVoyageLocation,
	},
	{
		Comment:       "Hunter fallback to voyage location gps must run before weather service",
		Issue:         "ENG-477",
		ShipID:        calc.HunterIdun,
		Start:         time.Time{}, // unbounded
		End:           time.Time{},
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.EnableFallbackToVoyageLocation,
	},
	{
		Comment:       "Hunter fallback to voyage location gps must run before weather service",
		Issue:         "ENG-477",
		ShipID:        calc.HunterLaga,
		Start:         time.Time{}, // unbounded
		End:           time.Time{},
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.EnableFallbackToVoyageLocation,
	},
	{
		Comment:       "Hunter fallback to voyage location gps must run before weather service",
		Issue:         "ENG-477",
		ShipID:        calc.HunterSaga,
		Start:         time.Time{}, // unbounded
		End:           time.Time{},
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.EnableFallbackToVoyageLocation,
	},
	{
		Comment:       "Bad Solomon Sea data point",
		Issue:         "ENG-461",
		ShipID:        calc.EpsSolomonSea,
		Start:         parseTime("2020-10-29 00:00"),
		End:           parseTime("2020-11-02 00:00"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc: func(pc calcapi.PropertyCalc) {
			l := labels.Noon(labels.ShaftSpeed)
			old := pc.GetNullableProperty(l)
			near := func(f1, f2 float64) bool {
				epsilon := .01
				return math.Abs(f1-f2) < epsilon
			}
			if old.Present() && near(old.Value(), 7534) {
				pc.SetProperty(l, 75.34)
			}
		},
	},

	{
		Comment:       "primary Lat/Lon is not reliable",
		Issue:         "ENG-614",
		ShipID:        calc.PacificBlue,
		Start:         time.Time{}, // unbounded
		End:           time.Time{},
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.EnableFallbackToVoyageLocation,
	},
	{
		Comment:       "primary Lat/Lon is not reliable",
		Issue:         "ENG-614",
		ShipID:        calc.PacificJade,
		Start:         time.Time{}, // unbounded
		End:           time.Time{},
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.EnableFallbackToVoyageLocation,
	},
	{
		Comment:       "resampling error for Voyage Location",
		Issue:         "ENG-756",
		ShipID:        calc.PacificBlue,
		Start:         parseTime("2021-01-25 20:00:00"),
		End:           parseTime("2021-01-25 22:00:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc: func(pc calcapi.PropertyCalc) {
			// simply copy last position
			if prev := pc.PreviousPosition(); prev != nil {
				pc.SetPosition(null.FloatFrom(prev.Latitude), null.FloatFrom(prev.Longitude))
			}
		},
	},
	{
		Comment:       "nulling out lat/lon data from reports prior to sensor data",
		Issue:         "ENG-615",
		ShipID:        calc.EpsQuebec,
		Start:         time.Time{}, // unbounded
		End:           parseTime("2021-01-15 02:00"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc:      RemoveBadGPS,
	},
	{
		Comment:       "interpolate gen fuel cons for erroneous data point",
		Issue:         "ENG-576",
		ShipID:        389, // Indian Solidarity
		Start:         parseTime("2021-01-03 04:00"),
		End:           parseTime("2021-01-03 06:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc: func(pc calcapi.PropertyCalc) {
			const genLshfoCons = "AE_LSFO_t_h"
			prevGenCons := pc.GetPreviousNullableProperty(genLshfoCons)
			nextGenCons := pc.GetNullablePropertyFromFeature(genLshfoCons, pc.FeatureIndex()+1)
			if prevGenCons.Present() && nextGenCons.Present() {
				currGenCons := models.SomeValue((prevGenCons.Value() + nextGenCons.Value()) / 2.0)
				pc.SetNullableProperty(genLshfoCons, currGenCons)
			}
		},
	},
	{
		Comment:       "alias SOG with Observed Speed for VO",
		Issue:         "ENG-634",
		ShipID:        calc.HunterFreya,
		Start:         parseTime("2020-10-23 17:00"),
		End:           time.Time{}, // if this sensor is fixed we can close the range
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc:      calc.AliasAndSmoothSOG,
	},
	{
		Comment:       "Fix Tenere STW tags",
		Issue:         "ENG-620",
		ShipID:        181,
		Start:         time.Time{},
		End:           parseTime("2020-12-19"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.Alias("NavigationThing_FilteredLogSpeed", labels.SpeedThroughWater),
	},
	{
		Comment:       "fallback on AIS",
		Issue:         "ENG-265",
		ShipID:        calc.EagleJay,
		Start:         time.Time{},
		End:           time.Time{},
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.EnableFallbackToAIS,
	},
	{
		Comment:       "fallback on AIS",
		Issue:         "ENG-676",
		ShipID:        calc.BulkFreedom,
		Start:         time.Time{},
		End:           time.Time{},
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.EnableFallbackToAIS,
	},
	{
		Comment:       "Fix Diamondway Erroneous Consumption",
		Issue:         "ENG-651",
		ShipID:        447,
		Start:         parseTime("2021-02-18 04:00:00"),
		End:           parseTime("2021-02-24 15:00:00"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc: func(pc calcapi.PropertyCalc) {
			pc.SetNullableProperty(labels.MainEngineFuelConsumption, pc.GetNullableProperty(labels.Noon(labels.MainEngineFuelConsumption)))
			pc.SetNullableProperty(labels.GeneratorFuelConsumption, pc.GetNullableProperty(labels.Noon(labels.GeneratorFuelConsumption)))
			pc.SetNullableProperty("Total Fuel Consumption", pc.GetNullableProperty(labels.Noon("Total Fuel Consumption")))
		},
	},
	{
		Comment:       "null out Pacific Jade data prior to March 18th",
		Issue:         "ENG-756",
		ShipID:        calc.PacificJade,
		Start:         time.Time{},
		End:           parseTime("2021-03-18"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CleanFunc: func(pc propertyClean) {
			NullAllFeatures(pc)
			RemoveBadGPS(pc)
		},
	},
	{
		Comment:       "null out Pacific Diamond data prior to March 20th",
		Issue:         "ENG-756",
		ShipID:        calc.EpsPacificDiamond,
		Start:         time.Time{},
		End:           parseTime("2021-03-20"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CleanFunc:     NullNoonFeatures,
	},
	{
		Comment:       "null out Vectis Progress Shaft Power",
		Issue:         "ENG-729",
		ShipID:        calc.VectisProgress,
		Start:         time.Time{},
		End:           parseTime("2021-03-01"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc:      calc.SetNull(labels.ShaftPower),
	},
	{
		Comment:       "Clear out fuel consumption before Jan 4th, 2021",
		Issue:         "ENG-769",
		ShipID:        calc.EpsIndianSolidarity,
		Start:         time.Time{},
		End:           parseTime("2021-01-04"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.SetNull(calc.EnamorFuelTags...),
	},
	{
		Comment:       "Remove Pacific Beryl noisy shaft power",
		Issue:         "ENG-792",
		ShipID:        calc.EpsPacificBeryl,
		Start:         parseTime("2021-01-12 00:00"),
		End:           parseTime("2021-03-30 00:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.SetNull(labels.ShaftPower),
	},
	{
		Comment:       "Remove FOC data for chevron asia energy",
		Issue:         "ENG-712",
		ShipID:        207,
		Start:         time.Time{},
		End:           parseTime("2021-01-28"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc:      calc.SetNull(calc.NS499FuelConsumptionTags...),
	},
	{
		Comment:       "Remove FOC data for chevron asia excellence",
		Issue:         "ENG-712",
		ShipID:        971,
		Start:         time.Time{},
		End:           parseTime("2021-01-12"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc:      calc.SetNull(calc.NS499FuelConsumptionTags...),
	},
	{
		Comment:       "Remove all data for chevron asia vision",
		Issue:         "ENG-712",
		ShipID:        896,
		Start:         time.Time{},
		End:           parseTime("2021-02-08"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc:      calc.SetNull(calc.NS499FuelConsumptionTags...),
	},
	{
		Comment:       "Remove bad STW sensor data for EPS Yukon",
		Issue:         "DMT-686",
		ShipID:        369116,
		Start:         parseTime("2021-01-04 00:00"),
		End:           parseTime("2021-03-22 00:00"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc:      calc.SetNull(labels.SpeedThroughWater),
	},
	{
		Comment:       "Remove Pacific Diamond bad data for rollout",
		Issue:         "DMT-649",
		ShipID:        calc.EpsPacificDiamond,
		Start:         time.Time{},
		End:           parseTime("2021-03-20 00:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CleanFunc: func(pc propertyClean) {
			NullAllFeatures(pc)
			RemoveBadGPS(pc)
		},
	},
	{
		Comment:       "Remove stw data for pacific gold",
		Issue:         "DMT-712",
		ShipID:        616,
		Start:         parseTime("2021-03-21 00:00"),
		End:           parseTime("2021-04-16 00:00"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc:      calc.SetNull(labels.SpeedThroughWater),
	},
	{
		Comment:       "Remove stw data for pacific gold",
		Issue:         "DMT-712",
		ShipID:        616,
		Start:         parseTime("2021-01-26 00:00"),
		End:           parseTime("2021-03-03 00:00"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc:      calc.SetNull(labels.SpeedThroughWater),
	},
	{
		Comment:       "Remove bad shaft speed and power for Diamondway",
		Issue:         "DMT-685",
		ShipID:        447,
		Start:         parseTime("2021-01-13 00:00"),
		End:           parseTime("2021-02-08 00:00"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc:      calc.SetNull(labels.ShaftSpeed, labels.ShaftPower),
	},
	{
		Comment:       "Remove erroneous GPS for Sunray",
		Issue:         "DPI-1287",
		ShipID:        283,
		Start:         parseTime("2021-03-04 22:00"),
		End:           parseTime("2021-03-05 15:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      RemoveBadGPS,
	},
	{
		Comment:       "Remove erroneous GPS for Sunray",
		Issue:         "DPI-1287",
		ShipID:        283,
		Start:         parseTime("2020-07-22 20:00"),
		End:           parseTime("2020-07-22 23:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      RemoveBadGPS,
	},
	{
		Comment:       "Remove erroneous GPS for Sunray",
		Issue:         "DPI-1287",
		ShipID:        283,
		Start:         parseTime("2020-07-12 07:00"),
		End:           parseTime("2020-07-12 09:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      RemoveBadGPS,
	},
	{
		Comment:       "Remove erroneous GPS for Sunray",
		Issue:         "DPI-1287",
		ShipID:        283,
		Start:         parseTime("2020-06-26 00:00"),
		End:           parseTime("2020-06-26 02:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      RemoveBadGPS,
	},
	{
		Comment:       "Remove erroneous GPS for Sunray",
		Issue:         "DPI-1287",
		ShipID:        283,
		Start:         parseTime("2020-06-25 20:00"),
		End:           parseTime("2020-06-25 22:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      RemoveBadGPS,
	},
	{
		Comment:       "Remove erroneous GPS for Sunray",
		Issue:         "DPI-1287",
		ShipID:        283,
		Start:         parseTime("2019-11-23 00:00"),
		End:           parseTime("2019-11-24 00:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      RemoveBadGPS,
	},
	{
		Comment:       "Remove erroneous GPS for Sunray",
		Issue:         "DPI-1287",
		ShipID:        283,
		Start:         parseTime("2019-11-17 17:00"),
		End:           parseTime("2019-11-17 19:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      RemoveBadGPS,
	},
	{
		Comment:       "Remove Shaft Power for EPS Irongate",
		Issue:         "DMT-684",
		ShipID:        1877,
		Start:         time.Time{},
		End:           parseTime("2021-03-06 00:00"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc:      calc.SetNull(labels.ShaftPower),
	},
	{
		Comment:       "Remove bad data for EPS CMA CGM PANAMA",
		Issue:         "DMT-743",
		ShipID:        110,
		Start:         parseTime("2021-04-04 00:00"),
		End:           parseTime("2021-04-28 00:00"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CleanFunc: func(pc propertyClean) {
			NullAllFeatures(pc)
			RemoveBadGPS(pc)
		},
	},
	{
		Comment:       "Round extremely small shaft values to zero for EPS Mount Bolivar",
		Issue:         "ENG-850",
		ShipID:        calc.EpsMountBolivar,
		Start:         time.Time{},
		End:           time.Time{},
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc: func(pc calcapi.PropertyCalc) {
			// Chosen because FE typically only shows 2 decimal places
			const epsilon = 0.001
			SetZeroIfWithinEpsilon(labels.ShaftSpeed, epsilon)(pc)
			SetZeroIfWithinEpsilon(labels.ShaftPower, epsilon)(pc)
		},
	},
	{
		Comment:       "EPS-Pacific-Cobalt-Remove-STW-sensor-data",
		Issue:         "DMT-783",
		ShipID:        146207,
		Start:         parseTime("2020-11-21 22:00"),
		End:           parseTime("2021-01-15 07:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.SetNull(labels.SpeedThroughWater),
	},
	{
		Comment:       "EPS-Pacific-Cobalt-Remove-STW-sensor-data",
		Issue:         "DMT-783",
		ShipID:        146207,
		Start:         parseTime("2021-01-29 08:00"),
		End:           parseTime("2021-02-25 08:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.SetNull(labels.SpeedThroughWater),
	},
	{
		Comment:       "Clean fuel data for Tyrrhenian Sea",
		Issue:         "DMT-827",
		ShipID:        89,
		Start:         time.Time{},
		End:           parseTime("2021-05-23 00:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.SetNull(labels.MainEngineFuelConsumption, labels.GeneratorFuelConsumption),
	},
	{
		Comment:       "Hunter Freya - Use Deprecated Fuel Tag Prior to New Tag Addition",
		Issue:         "ENG-1007",
		ShipID:        calc.HunterFreya,
		Start:         time.Time{}, // Unbounded
		End:           parseTime("2021-05-27 00:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc: func(pc calcapi.PropertyCalc) {
			// Sets the fuel consumption tag to the only tag previously available for these two ships to the standard tag and prevents double counting for the remainder of the data
			pc.SetNullableProperty("Engines_Main_FO_Flow", pc.GetNullableProperty("Engines_Main_1_Fuel_Oil_System_FO_Inlet_Flow_Mass"))
		},
	},
	{
		Comment:       "Hunter Frigg - Use Deprecated Fuel Tag Prior to New Tag Addition",
		Issue:         "ENG-1007",
		ShipID:        calc.HunterFrigg,
		Start:         time.Time{}, // Unbounded
		End:           parseTime("2021-05-27 00:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc: func(pc calcapi.PropertyCalc) {
			// Sets the fuel consumption tag to the only tag previously available for these two ships to the standard tag and prevents double counting for the remainder of the data
			pc.SetNullableProperty("Engines_Main_FO_Flow", pc.GetNullableProperty("Engines_Main_1_Fuel_Oil_System_FO_Inlet_Flow_Mass"))
		},
	},
	{
		Comment:       "EPS - Pacific Gold weird spike",
		Issue:         "ENG-924",
		ShipID:        calc.PacificGold,
		Start:         parseTime("2021-07-06 06:00"),
		End:           parseTime("2021-07-06 21:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      RemoveBadGPS,
	},
	{
		Comment:       "EPS - Pacific Gold Weird GPS Spike 2",
		Issue:         "DPI-1678",
		ShipID:        calc.PacificGold,
		Start:         parseTime("2022-05-03 15:00"),
		End:           parseTime("2022-05-04 05:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      RemoveBadGPS,
	},
	{
		Comment:       "Clean fuel data for Nordic Orion",
		Issue:         "ENG-1105",
		ShipID:        calc.NordicOrion,
		Start:         parseTime("2021-07-07 06:00"),
		End:           parseTime("2021-09-08 00:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.SetNull("G/E Outlet Mass Flow (MT/hr)", "G/E Inlet Mass Flow (MT/hr)", "M/E Mass Flow (MT/hr)"),
	},
	{
		Comment:       "Clean fuel data for Nordic Olympic",
		Issue:         "ENG-1105",
		ShipID:        calc.NordicOlympic,
		Start:         parseTime("2021-11-04 20:00"),
		End:           parseTime("2021-12-14 00:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.SetNull("G/E Outlet Mass Flow (MT/hr)", "G/E Inlet Mass Flow (MT/hr)", "M/E Mass Flow (MT/hr)"),
	},
	{
		Comment:       "Clean Shaft Power data for Bulk Destiny",
		Issue:         "ENG-1124",
		ShipID:        calc.BulkDestiny,
		Start:         parseTime("2021-09-07 01:00"),
		End:           parseTime("2022-01-22 00:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.SetNull(labels.ShaftPower, labels.ShaftTorque),
	},
	{
		Comment:       "Naively Forward Fill MEFC",
		Issue:         "VOTR-10",
		ShipID:        calc.PdSana,
		Start:         parseTime("2022-04-20 04:00:00"),
		End:           time.Time{}, // Unbounded
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc: func(pc calcapi.PropertyCalc) {
			pc.SetNullableProperty(labels.MainEngineFuelConsumption, models.SomeValue(0.1))
			pc.SetNullableProperty(labels.GeneratorFuelConsumption, models.SomeValue(0.1))
			pc.SetNullableProperty("Total Fuel Consumption", models.SomeValue(0.1))
		},
	},
	{
		Comment:       "Onboard AIS Fallback for BW Brussels",
		Issue:         "DPI-1680",
		ShipID:        351, //BW Brussels
		Start:         time.Time{},
		End:           parseTime("2022-05-31 13:00:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.FallbackForZeroPosition("H2259.AIS_Latitude", "H2259.AIS_Longitude"),
	},
	{
		Comment:       "Onboard AIS Fallback for BW Brussels",
		Issue:         "DPI-1680",
		ShipID:        351, //BW Brussels
		Start:         time.Time{},
		End:           parseTime("2022-05-31 13:00:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.FallbackForZeroPosition("H2259.AIS_Latitude", "H2259.AIS_Longitude"),
	},
	{
		Comment:       "Spire AIS Fallback for BW Brussels",
		Issue:         "DPI-1680",
		ShipID:        351, //BW Brussels
		Start:         parseTime("2022-05-31 12:00:00"),
		End:           time.Time{},
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.UseAISforPosition(labels.AisLatitude, labels.AisLongitude),
	},
	{
		Comment:       "Removed brussel bad GPS in June 2021",
		Issue:         "DPI-1680",
		ShipID:        351, //BW Brussel
		Start:         parseTime("2021-06-21 23:00:00"),
		End:           parseTime("2021-06-24 11:00:00"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc: func(pc calcapi.PropertyCalc) {
			pc.SetPosition(null.Float{}, null.Float{})
		},
	},
	{
		Comment:       "Remove Lake Wanaka shaft power",
		Issue:         "ENG-1263",
		ShipID:        612, //Lake Wanaka
		Start:         parseTime("2021-11-26 07:00"),
		End:           parseTime("2021-11-27 12:00"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc:      calc.SetNull(labels.ShaftPower),
	},
	{
		Comment:       "Remove Lake Wanaka shaft power",
		Issue:         "ENG-1263",
		ShipID:        612, //Lake Wanaka
		Start:         parseTime("2022-03-06 11:00"),
		End:           parseTime("2022-03-06 22:00"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc:      calc.SetNull(labels.ShaftSpeed),
	},
	{
		Comment: " Remove Bad Data Points",
		Issue: "ENG-1263",
		ShipID: 612, //Lake Wanaka
		Start: parseTime("2021-09-25 16:00"),
		End: parseTime("2021-09-26 12:00"),
		Unconditional: true,
		Stage: PostVesselAnatomyStage,
		CalcFunc: calc.SetNull(labels.MainEngineFuelConsumption),
	},
	{
		Comment: " Remove Bad Data Points",
		Issue: "ENG-1263",
		ShipID: 612, //Lake Wanaka
		Start: parseTime("2021-09-25 16:00"),
		End: parseTime("2021-09-26 12:00"),
		Unconditional: true,
		Stage: PostVesselAnatomyStage,
		CalcFunc: calc.SetNull("Total Fuel Consumption"),
	},
	{
		Comment:       "MEFC fallback",
		Issue:         "VOTR-85",
		ShipID:        447,
		Start:         parseTime("2022-06-09 07:00"),
		End:           parseTime("2022-06-16"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc: calc.Alias(labels.NoonMainEngineFuelConsumption, labels.MainEngineFuelConsumption),
	},
	{
		Comment:       "MEFC fallback",
		Issue:         "VOTR-85",
		ShipID:        447,
		Start:         parseTime("2022-06-09 07:00"),
		End:           parseTime("2022-06-16"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc: calc.Alias(labels.Noon(labels.System.TOTAL.Consumption(labels.Fuel.HFO)), labels.System.TOTAL.Consumption(labels.Fuel.HFO)),
	},
	{
		Comment:       "MEFC fallback",
		Issue:         "VOTR-85",
		ShipID:        447,
		Start:         parseTime("2022-06-09 07:00"),
		End:           parseTime("2022-06-16"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc: calc.Alias(labels.Noon(labels.System.TOTAL.FuelConsumption()), labels.System.TOTAL.FuelConsumption()),
	},
	{
		Comment:       "MEFC fallback",
		Issue:         "VOTR-85",
		ShipID:        447,
		Start:         parseTime("2022-06-09 07:00"),
		End:           parseTime("2022-06-16"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc: calc.Alias(labels.Noon(labels.System.ME.Consumption(labels.Fuel.HFO)), labels.System.ME.Consumption(labels.Fuel.HFO)),
	},
	{
		Comment:       "Clean fuel data for Coral EnergICE",
		Issue:         "DPI-1833",
		ShipID:        759,
		Start:         time.Time{},
		End:           parseTime("2022-07-21 06:00"), // Unbounded
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc: calc.SetNull("AUX ENG 3 GAS FLOW METER V", "FUEL GAS FLOW THERMAL OIL BOILER V"),
	},
	{
		Comment:       "Clean fuel data for Coral EnergICE",
		Issue:         "DPI-1833",
		ShipID:        759,
		Start:         parseTime("2022-05-04 02:00"),
		End:           parseTime("2022-05-05 16:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc: calc.SetNull("AUX ENG 2 GAS FLOW METER V"),
	},
	{
		Comment:       "Clean fuel data for Coral EnergICE",
		Issue:         "DPI-1833",
		ShipID:        759,
		Start:         parseTime("2022-05-15 03:00"),
		End:           parseTime("2022-05-25 07:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc: calc.SetNull("AUX ENG 2 GAS FLOW METER V"),
	},
	{
		Comment:       "Clean fuel data for Coral EnergICE",
		Issue:         "DPI-1833",
		ShipID:        759,
		Start:         parseTime("2022-06-01 03:00"),
		End:           parseTime("2022-06-03 19:00"),
		Unconditional: true,
		Stage:         PreVesselAnatomyStage,
		CalcFunc:      calc.SetNull("AUX ENG 2 GAS FLOW METER V"),
	},
	{
		Comment:       "eps-fairway-remove-change-bad-data",
		Issue:         "ENG-1340",
		ShipID:        119,
		Start:         parseTime("2022-04-28 00:00"),
		End:           parseTime("2022-04-28 17:00"),
		Unconditional: true,
		Stage:         PostVesselAnatomyStage,
		CalcFunc: func(pc calcapi.PropertyCalc) {	
			pc.SetProperty("(Noon) Longitude", pc.GetProperty("(Noon) Longitude")*-1)		
			//pc.SetNullableProperty("(Noon) Longitude", null.FloatFrom(-1*pc.GetNullableProperty("(Noon) Longitude")))
		},
		
	},
}
