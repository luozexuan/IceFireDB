// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package webrtc

import (
	"fmt"

	"github.com/pion/ice/v4"
)

// ICECandidate represents a ice candidate
type ICECandidate struct {
	statsID        string
	Foundation     string           `json:"foundation"`
	Priority       uint32           `json:"priority"`
	Address        string           `json:"address"`
	Protocol       ICEProtocol      `json:"protocol"`
	Port           uint16           `json:"port"`
	Typ            ICECandidateType `json:"type"`
	Component      uint16           `json:"component"`
	RelatedAddress string           `json:"relatedAddress"`
	RelatedPort    uint16           `json:"relatedPort"`
	TCPType        string           `json:"tcpType"`
	SDPMid         string           `json:"sdpMid"`
	SDPMLineIndex  uint16           `json:"sdpMLineIndex"`
}

// Conversion for package ice

func newICECandidatesFromICE(iceCandidates []ice.Candidate, sdpMid string, sdpMLineIndex uint16) ([]ICECandidate, error) {
	candidates := []ICECandidate{}

	for _, i := range iceCandidates {
		c, err := newICECandidateFromICE(i, sdpMid, sdpMLineIndex)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, c)
	}

	return candidates, nil
}

func newICECandidateFromICE(i ice.Candidate, sdpMid string, sdpMLineIndex uint16) (ICECandidate, error) {
	typ, err := convertTypeFromICE(i.Type())
	if err != nil {
		return ICECandidate{}, err
	}
	protocol, err := NewICEProtocol(i.NetworkType().NetworkShort())
	if err != nil {
		return ICECandidate{}, err
	}

	c := ICECandidate{
		statsID:       i.ID(),
		Foundation:    i.Foundation(),
		Priority:      i.Priority(),
		Address:       i.Address(),
		Protocol:      protocol,
		Port:          uint16(i.Port()),
		Component:     i.Component(),
		Typ:           typ,
		TCPType:       i.TCPType().String(),
		SDPMid:        sdpMid,
		SDPMLineIndex: sdpMLineIndex,
	}

	if i.RelatedAddress() != nil {
		c.RelatedAddress = i.RelatedAddress().Address
		c.RelatedPort = uint16(i.RelatedAddress().Port)
	}

	return c, nil
}

func (c ICECandidate) toICE() (ice.Candidate, error) {
	candidateID := c.statsID
	switch c.Typ {
	case ICECandidateTypeHost:
		config := ice.CandidateHostConfig{
			CandidateID: candidateID,
			Network:     c.Protocol.String(),
			Address:     c.Address,
			Port:        int(c.Port),
			Component:   c.Component,
			TCPType:     ice.NewTCPType(c.TCPType),
			Foundation:  c.Foundation,
			Priority:    c.Priority,
		}
		return ice.NewCandidateHost(&config)
	case ICECandidateTypeSrflx:
		config := ice.CandidateServerReflexiveConfig{
			CandidateID: candidateID,
			Network:     c.Protocol.String(),
			Address:     c.Address,
			Port:        int(c.Port),
			Component:   c.Component,
			Foundation:  c.Foundation,
			Priority:    c.Priority,
			RelAddr:     c.RelatedAddress,
			RelPort:     int(c.RelatedPort),
		}
		return ice.NewCandidateServerReflexive(&config)
	case ICECandidateTypePrflx:
		config := ice.CandidatePeerReflexiveConfig{
			CandidateID: candidateID,
			Network:     c.Protocol.String(),
			Address:     c.Address,
			Port:        int(c.Port),
			Component:   c.Component,
			Foundation:  c.Foundation,
			Priority:    c.Priority,
			RelAddr:     c.RelatedAddress,
			RelPort:     int(c.RelatedPort),
		}
		return ice.NewCandidatePeerReflexive(&config)
	case ICECandidateTypeRelay:
		config := ice.CandidateRelayConfig{
			CandidateID: candidateID,
			Network:     c.Protocol.String(),
			Address:     c.Address,
			Port:        int(c.Port),
			Component:   c.Component,
			Foundation:  c.Foundation,
			Priority:    c.Priority,
			RelAddr:     c.RelatedAddress,
			RelPort:     int(c.RelatedPort),
		}
		return ice.NewCandidateRelay(&config)
	default:
		return nil, fmt.Errorf("%w: %s", errICECandidateTypeUnknown, c.Typ)
	}
}

func convertTypeFromICE(t ice.CandidateType) (ICECandidateType, error) {
	switch t {
	case ice.CandidateTypeHost:
		return ICECandidateTypeHost, nil
	case ice.CandidateTypeServerReflexive:
		return ICECandidateTypeSrflx, nil
	case ice.CandidateTypePeerReflexive:
		return ICECandidateTypePrflx, nil
	case ice.CandidateTypeRelay:
		return ICECandidateTypeRelay, nil
	default:
		return ICECandidateType(t), fmt.Errorf("%w: %s", errICECandidateTypeUnknown, t)
	}
}

func (c ICECandidate) String() string {
	ic, err := c.toICE()
	if err != nil {
		return fmt.Sprintf("%#v failed to convert to ICE: %s", c, err)
	}
	return ic.String()
}

// ToJSON returns an ICECandidateInit
// as indicated by the spec https://w3c.github.io/webrtc-pc/#dom-rtcicecandidate-tojson
func (c ICECandidate) ToJSON() ICECandidateInit {
	candidateStr := ""

	candidate, err := c.toICE()
	if err == nil {
		candidateStr = candidate.Marshal()
	}

	return ICECandidateInit{
		Candidate:     fmt.Sprintf("candidate:%s", candidateStr),
		SDPMid:        &c.SDPMid,
		SDPMLineIndex: &c.SDPMLineIndex,
	}
}
