// Package docs GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag
package docs

import (
	"bytes"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/": {
            "get": {
                "description": "get the status of server.",
                "consumes": [
                    "*/*"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "root"
                ],
                "summary": "Show the status of server.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    }
                }
            }
        },
        "/v1/events": {
            "get": {
                "description": "Lists the user's events in reverse chronological order",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "events"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "DIMO device ID",
                        "name": "device_id",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Event type, usually User or Device",
                        "name": "type",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Further refinement of the event type",
                        "name": "sub_type",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/controllers.EventResponseEntry"
                            }
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "controllers.EventResponseEntry": {
            "type": "object",
            "properties": {
                "data": {
                    "description": "Data is an event-specific object containing more information about the event."
                },
                "deviceId": {
                    "description": "DeviceID is the DIMO device ID when the event relates to a device.",
                    "type": "string"
                },
                "id": {
                    "description": "ID is the event ID, a KSUID.",
                    "type": "string"
                },
                "subType": {
                    "description": "SubType is a more specific classification of the event.",
                    "type": "string",
                    "example": "Created"
                },
                "timestamp": {
                    "type": "string"
                },
                "type": {
                    "description": "Type is the type of the event, either \"User\" or \"Device\".",
                    "type": "string",
                    "example": "Device"
                },
                "userId": {
                    "description": "UserID is the DIMO user id coming from the authentication service.",
                    "type": "string"
                }
            }
        }
    },
    "securityDefinitions": {
        "BearerAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "1.0",
	Host:        "",
	BasePath:    "",
	Schemes:     []string{},
	Title:       "DIMO Events API",
	Description: "",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
		"escape": func(v interface{}) string {
			// escape tabs
			str := strings.Replace(v.(string), "\t", "\\t", -1)
			// replace " with \", and if that results in \\", replace that with \\\"
			str = strings.Replace(str, "\"", "\\\"", -1)
			return strings.Replace(str, "\\\\\"", "\\\\\\\"", -1)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register("swagger", &s{})
}