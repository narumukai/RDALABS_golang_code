package cleanup

import (
	"fmt"
	"github.com/jinzhu/now"
	"github.com/nautiluslabsco/ln/features/calc"
	"github.com/nautiluslabsco/ln/features/calc/calcapi"
	"github.com/nautiluslabsco/ln/shared/constants/labels"
	"github.com/nautiluslabsco/ln/shared/constants/units"
	"github.com/nautiluslabsco/ln/shared/models"
	"github.com/nautiluslabsco/ln/shared/nmath"
	"github.com/nautiluslabsco/null"
	log "github.com/sirupsen/logrus"
	"math"
	"time"
)

type Stage string

const (
	PreVesselAnatomyStage  Stage = "pre-vessel-anatomy"
	PostVesselAnatomyStage Stage = "post-vessel-anatomy"
)

type CleanupFunc struct {
	Comment       string
	Issue         string
	ShipID        int64
	Start         time.Time
	End           time.Time
	Unconditional bool
	Stage         Stage
	CalcFunc      func(calcapi.PropertyCalc)
	CleanFunc     func(propertyClean)
}

// PropertyCalc interface, but with some nastier
// functions to clean up data
type propertyClean interface {
	calcapi.PropertyCalc
	NullAllProperties()
	NullPrefixedProperties(prefix string)
}

func CleanupFuncs(stage Stage, onlyUnconditional bool) []func(calcapi.PropertyCalc) {
	var calcFuncsToRun []func(calcapi.PropertyCalc)
	for _, cleanupFunc := range cleanupFuncs {
		if onlyUnconditional && !cleanupFunc.Unconditional || stage != cleanupFunc.Stage {
			continue
		}
		calcFuncsToRun = append(calcFuncsToRun, cleanupFunc.calcFunc())
	}
	return calcFuncsToRun
}

func parseTime(t string) time.Time {
	tp, err := now.ParseInLocation(time.UTC, t)
	if err != nil {
		panic(err)
	}
	return tp
}

func (cfunc CleanupFunc) calcFunc() func(calcapi.PropertyCalc) {
	if cfunc.Issue == "" {
		panic("No NAUT issue associated with cleanup calc func")
	}
	return func(pc calcapi.PropertyCalc) {
		if pc.GetShip().ID != cfunc.ShipID {
			return
		}

		endTime := cfunc.End
		if endTime.IsZero() {
			endTime = time.Now()
		}

		if cfunc.Start.Before(pc.Time()) && endTime.After(pc.Time()) {
			log.Debugf("Cleaning up data on %s for ship %d because of %s", pc.Time().Format(time.RFC3339), pc.GetShip().ID, cfunc.Issue)
			if cfunc.CalcFunc != nil {
				cfunc.CalcFunc(pc)
			}

			if cfunc.CleanFunc != nil {
				if c, ok := pc.(propertyClean); ok {
					cfunc.CleanFunc(c)
				} else {
					log.Debugf("Somehow unable to convert to propertyClean type")
				}
			}
		}
	}
}

func copernicusModeledSTW(pc calcapi.PropertyCalcGetter) null.Float {
	coeffs := []float64{
		0.191637416,
		-0.153619789,
		-0.123217962,
		-0.324622524,
		0.057272980,
		0.044914439,
		-0.035553420,
		-0.000633009,
	}

	compositeWind := nmath.ScalarProjection(
		pc.GetProperty(string(calc.WS_TrueWindSpeed)),
		pc.GetProperty(string(calc.WS_TrueWindDir)),
		pc.GetProperty(labels.Heading))
	compositeCurrent := nmath.ScalarProjection(
		pc.GetProperty(string(calc.WS_SeaCurSpeed)),
		pc.GetProperty(string(calc.WS_SeaCurDir)),
		pc.GetProperty(labels.Heading))

	modeled := coeffs[0]*pc.GetProperty(labels.ShaftSpeed) +
		coeffs[1]*pc.GetProperty(labels.DraftAft) +
		coeffs[2]*pc.GetProperty(labels.Trim) +
		coeffs[3]*pc.GetProperty(string(calc.WS_SigWaveHeight)) +
		coeffs[4]*compositeCurrent*compositeCurrent +
		coeffs[5]*compositeCurrent +
		coeffs[6]*compositeWind +
		coeffs[7]*compositeWind*compositeWind +
		1.4246

	if pc.HasError() {
		return null.Float{}
	}

	return null.FloatFrom(modeled)
}

func OverrideLatLonSign(pc calcapi.PropertyCalc) {
	noonLat := pc.GetNullableProperty("(Noon) Latitude")
	noonLon := pc.GetNullableProperty("(Noon) Longitude")
	pos := pc.Position()
	//fmt.print("testing", pos)
	if noonLat.Absent() || noonLon.Absent() || pos == nil {
		return
	}

	if noonLat.Value() < 0 && pos.Latitude > 0.0 {
		pos.Latitude *= -1
	}
	if noonLon.Value() < 0 && pos.Longitude > 0.0 {
		pos.Longitude *= -1
	}
	pc.SetPosition(null.FloatFrom(pos.Latitude), null.FloatFrom(pos.Longitude))
}

func OverrideChevronGeneratorPower(pc calcapi.PropertyCalc) {
	for i := 1; i < 5; i++ {
		mgPower := pc.GetNullableProperty(fmt.Sprintf("M/G %d Power", i))
		if mgPower.Present() {
			pc.SetPropertyWithUnit(fmt.Sprintf("Generator %d Power", i), mgPower.Value(), units.KiloWatts.String())
		}
	}
}

func NegateLatitude(pc calcapi.PropertyCalc) {
	pos := pc.Position()	
	if pos == nil {
		return
	}	
	pc.SetPosition(null.FloatFrom(-pos.Latitude), null.FloatFrom(pos.Longitude))
}
func NegateLongitude(pc calcapi.PropertyCalc) {
	pos := pc.Position()
	//pc.SetNullableProperty("(Noon) Longitude", null.FloatFrom(-1*pc.GetNullableProperty("(Noon) Longitude")))
	pc.SetPosition(null.FloatFrom(pos.Longitude), null.FloatFrom(-pos.Longitude))
}


func RemoveBadGPS(pc calcapi.PropertyCalc) {
	pc.SetPosition(null.Float{}, null.Float{})
	
}

func NullNoonFeatures(pc propertyClean) {
	pc.NullPrefixedProperties(labels.Noon(""))
}

func NullAllFeatures(pc propertyClean) {
	pc.NullAllProperties()
}

func SetZeroIfWithinEpsilon(label string, epsilon float64) func(propertyCalc calcapi.PropertyCalc) {
	return func(pc calcapi.PropertyCalc) {
		if v := pc.GetNullableProperty(label); v.Present() && math.Abs(v.Value()) < epsilon {
			pc.SetNullableProperty(label, models.SomeValue(0))
		}
	}
}
