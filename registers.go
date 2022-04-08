package main

import (
	"encoding/binary"
	"math"
)

type register struct {
	addr      uint16
	size      uint16
	converter valueConverter
}
type registers map[string]registerToSensor

type sensor struct {
	stateClass string // measurement or total
	unit       string // kWh, W, A
}

func measurementSensor(unit string) *sensor {
	return &sensor{"measurement", unit}
}

func totalSensor(unit string) *sensor {
	return &sensor{"total", unit}
}

type registerToSensor struct {
	reg register
	s   *sensor
}

type valueConverter func([]byte) interface{}

// float32Converter covert bytes into big endian float32
func float32Converter(bytes []byte) interface{} {
	bits := binary.BigEndian.Uint32(bytes)
	float := math.Float32frombits(bits)
	return float
}

// uint16Converter covert bytes into big endian  uint16
func uint16Converter(bytes []byte) interface{} {
	return binary.BigEndian.Uint16(bytes)
}

// uint32Converter covert bytes into big endian uint32
func uint32Converter(bytes []byte) interface{} {
	return binary.BigEndian.Uint32(bytes)
}

// float32FromUint32With3DecimalConverter converts 2 decimal digits value represent by big endian uint16 into float32
// this is exclusive for OR-WE-514 power meter
func float32FromUint16DecimalConverter(b []byte) interface{} {
	return float32(uint16Converter(b).(uint16)) / 100
}

// float32FromUint32With3DecimalConverter converts 2 decimal digits value represent by big endian uint32 into float32
// this is exclusive for OR-WE-514 power meter
func float32FromUint32With2DecimalConverter(b []byte) interface{} {
	return float32(uint32Converter(b).(uint32)) / 100
}

// float32FromUint32With3DecimalConverter converts 3 decimal digits value represent by big endian uint32 into float32
// this is exclusive for OR-WE-514 power meter
func float32FromUint32With3DecimalConverter(b []byte) interface{} {
	return float32(uint32Converter(b).(uint32)) / 1000
}

var registers514 = registers{
	"Grid Frequency":        {register{0x0130, 1, float32FromUint16DecimalConverter}, measurementSensor("Hz")},      // 50 01
	"Voltage":               {register{0x0131, 1, float32FromUint16DecimalConverter}, measurementSensor("V")},       // 228 76
	"Current":               {register{0x0139, 2, float32FromUint32With3DecimalConverter}, measurementSensor("A")},  // 5 472
	"Total Active Power":    {register{0x0140, 2, float32FromUint32With3DecimalConverter}, measurementSensor("kW")}, // 1 254
	"Total reactive power":  {register{0x0148, 2, float32FromUint32With3DecimalConverter}, measurementSensor("kW")},
	"Total Apparent Power":  {register{0x0150, 2, float32FromUint32With3DecimalConverter}, measurementSensor("kW")},
	"Total Power Factor":    {register{0x0158, 1, float32FromUint16DecimalConverter}, measurementSensor("kW")},
	"Total Active Energy":   {register{0xA000, 2, float32FromUint32With2DecimalConverter}, totalSensor("kWh")}, // 6734 69
	"Total Reactive Energy": {register{0xA01E, 2, float32FromUint32With2DecimalConverter}, totalSensor("kWh")},
}

var registers517 = registers{
	"L1 Voltage":                 {register{0x000E, 2, float32Converter}, measurementSensor("V")},
	"L2 Voltage":                 {register{0x0010, 2, float32Converter}, measurementSensor("V")},
	"L3 Voltage":                 {register{0x0012, 2, float32Converter}, measurementSensor("V")},
	"Grid Frequency":             {register{0x0014, 2, float32Converter}, measurementSensor("Hz")},
	"L1 Current":                 {register{0x0016, 2, float32Converter}, measurementSensor("A")},
	"L2 Current":                 {register{0x0018, 2, float32Converter}, measurementSensor("A")},
	"L3 Current":                 {register{0x001A, 2, float32Converter}, measurementSensor("A")},
	"Total Active Power":         {register{0x001C, 2, float32Converter}, measurementSensor("kW")},
	"L1 Active Power":            {register{0x001E, 2, float32Converter}, measurementSensor("kW")},
	"L2 Active Power":            {register{0x0020, 2, float32Converter}, measurementSensor("kW")},
	"L3 Active Power":            {register{0x0022, 2, float32Converter}, measurementSensor("kW")},
	"Total reactive power":       {register{0x0024, 2, float32Converter}, measurementSensor("kW")},
	"L1 reactive power":          {register{0x0026, 2, float32Converter}, measurementSensor("kW")},
	"L2 reactive power":          {register{0x0028, 2, float32Converter}, measurementSensor("kW")},
	"L3 reactive power":          {register{0x002A, 2, float32Converter}, measurementSensor("kW")},
	"Total Apparent Power":       {register{0x002C, 2, float32Converter}, measurementSensor("kW")},
	"L1 Apparent Power":          {register{0x002E, 2, float32Converter}, measurementSensor("kW")},
	"L2 Apparent Power":          {register{0x0030, 2, float32Converter}, measurementSensor("kW")},
	"L3 Apparent Power":          {register{0x0032, 2, float32Converter}, measurementSensor("kW")},
	"Total Power Factor":         {register{0x0034, 2, float32Converter}, measurementSensor("kW")},
	"L1 Power Factor":            {register{0x0036, 2, float32Converter}, measurementSensor("kW")},
	"L2 Power Factor":            {register{0x0038, 2, float32Converter}, measurementSensor("kW")},
	"L3 Power Factor":            {register{0x003A, 2, float32Converter}, measurementSensor("kW")},
	"Total Active Energy":        {register{0x0100, 2, float32Converter}, totalSensor("kWh")},
	"L1 Total Active Energy":     {register{0x0102, 2, float32Converter}, totalSensor("kWh")},
	"L2 Total Active Energy":     {register{0x0104, 2, float32Converter}, totalSensor("kWh")},
	"L3 Total Active Energy":     {register{0x0106, 2, float32Converter}, totalSensor("kWh")},
	"Forward Active Energy":      {register{0x0108, 2, float32Converter}, totalSensor("kWh")},
	"L1 Forward Active Energy":   {register{0x010A, 2, float32Converter}, totalSensor("kWh")},
	"L2 Forward Active Energy":   {register{0x010C, 2, float32Converter}, totalSensor("kWh")},
	"L3 Forward Active Energy":   {register{0x010E, 2, float32Converter}, totalSensor("kWh")},
	"Reverse Active Energy":      {register{0x0110, 2, float32Converter}, totalSensor("kWh")},
	"L1 Reverse Active Energy":   {register{0x0112, 2, float32Converter}, totalSensor("kWh")},
	"L2 Reverse Active Energy":   {register{0x0114, 2, float32Converter}, totalSensor("kWh")},
	"L3 Reverse Active Energy":   {register{0x0116, 2, float32Converter}, totalSensor("kWh")},
	"Total Reactive Energy":      {register{0x0118, 2, float32Converter}, totalSensor("kWh")},
	"L1 Reactive Energy":         {register{0x011A, 2, float32Converter}, totalSensor("kWh")},
	"L2 Reactive Energy":         {register{0x011C, 2, float32Converter}, totalSensor("kWh")},
	"L3 Reactive Energy":         {register{0x011E, 2, float32Converter}, totalSensor("kWh")},
	"Forward Reactive Energy":    {register{0x0120, 2, float32Converter}, totalSensor("kWh")},
	"L1 Forward Reactive Energy": {register{0x0122, 2, float32Converter}, totalSensor("kWh")},
	"L2 Forward Reactive Energy": {register{0x0124, 2, float32Converter}, totalSensor("kWh")},
	"L3 Forward Reactive Energy": {register{0x0126, 2, float32Converter}, totalSensor("kWh")},
	"Reverse Reactive Energy":    {register{0x0128, 2, float32Converter}, totalSensor("kWh")},
	"L1 Reverse Reactive Energy": {register{0x012A, 2, float32Converter}, totalSensor("kWh")},
	"L2 Reverse Reactive Energy": {register{0x012C, 2, float32Converter}, totalSensor("kWh")},
	"L3 Reverse Reactive Energy": {register{0x012E, 2, float32Converter}, totalSensor("kWh")},
	"T1 Total Active energy":     {register{0x0130, 2, float32Converter}, totalSensor("kWh")},
	"T1 Forward Active Energy":   {register{0x0132, 2, float32Converter}, totalSensor("kWh")},
	"T1 Reverse Active Energy":   {register{0x0134, 2, float32Converter}, totalSensor("kWh")},
	"T1 Total Reactive Energy":   {register{0x0136, 2, float32Converter}, totalSensor("kWh")},
	"T1 Forward Reactive Energy": {register{0x0138, 2, float32Converter}, totalSensor("kWh")},
	"T1 Reverse Reactive Energy": {register{0x013A, 2, float32Converter}, totalSensor("kWh")},
	"T2 Total Active energy":     {register{0x013C, 2, float32Converter}, totalSensor("kWh")},
	"T2 Forward Active Energy":   {register{0x013E, 2, float32Converter}, totalSensor("kWh")},
	"T2 Reverse Active Energy":   {register{0x0140, 2, float32Converter}, totalSensor("kWh")},
	"T2 Total Reactive Energy":   {register{0x0142, 2, float32Converter}, totalSensor("kWh")},
	"T2 Forward Reactive Energy": {register{0x0144, 2, float32Converter}, totalSensor("kWh")},
	"T2 Reverse Reactive Energy": {register{0x0146, 2, float32Converter}, totalSensor("kWh")},
	"T3 Total Active energy":     {register{0x0148, 2, float32Converter}, totalSensor("kWh")},
	"T3 Forward Active Energy":   {register{0x014A, 2, float32Converter}, totalSensor("kWh")},
	"T3 Reverse Active Energy":   {register{0x014C, 2, float32Converter}, totalSensor("kWh")},
	"T3 Total Reactive Energy":   {register{0x014E, 2, float32Converter}, totalSensor("kWh")},
	"T3 Forward Reactive Energy": {register{0x0150, 2, float32Converter}, totalSensor("kWh")},
	"T3 Reverse Reactive Energy": {register{0x0152, 2, float32Converter}, totalSensor("kWh")},
	"T4 Total Active energy":     {register{0x0154, 2, float32Converter}, totalSensor("kWh")},
	"T4 Forward Active Energy":   {register{0x0156, 2, float32Converter}, totalSensor("kWh")},
	"T4 Reverse Active Energy":   {register{0x0158, 2, float32Converter}, totalSensor("kWh")},
	"T4 Total Reactive Energy":   {register{0x015A, 2, float32Converter}, totalSensor("kWh")},
	"T4 Forward Reactive Energy": {register{0x015C, 2, float32Converter}, totalSensor("kWh")},
	"T4 Reverse Reactive Energy": {register{0x015E, 2, float32Converter}, totalSensor("kWh")},
}
