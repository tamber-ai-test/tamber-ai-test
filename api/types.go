// Code generated from JSON Schema using quicktype. DO NOT EDIT.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    types, err := UnmarshalTypes(bytes)
//    bytes, err = types.Marshal()

package main

import "encoding/json"

func UnmarshalTypes(data []byte) (Types, error) {
	var r Types
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Types) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Types struct {
	Schema      string          `json:"$schema"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Type        string          `json:"type"`
	Properties  TypesProperties `json:"properties"`
	Required    []string        `json:"required"`
}

type TypesProperties struct {
	ID         ID         `json:"id"`
	Name       ID         `json:"name"`
	Email      Email      `json:"email"`
	IsVerified IsVerified `json:"isVerified"`
	Address    Address    `json:"address"`
	Tags       Tags       `json:"tags"`
}

type Address struct {
	Description string            `json:"description"`
	Type        string            `json:"type"`
	Properties  AddressProperties `json:"properties"`
	Required    []string          `json:"required"`
}

type AddressProperties struct {
	Street  Items `json:"street"`
	City    Items `json:"city"`
	ZipCode Items `json:"zipCode"`
}

type Items struct {
	Type string `json:"type"`
}

type Email struct {
	Description string `json:"description"`
	Type        string `json:"type"`
	Format      string `json:"format"`
}

type ID struct {
	Description string `json:"description"`
	Type        string `json:"type"`
}

type IsVerified struct {
	Description string `json:"description"`
	Type        string `json:"type"`
	Default     bool   `json:"default"`
}

type Tags struct {
	Description string `json:"description"`
	Type        string `json:"type"`
	Items       Items  `json:"items"`
}
