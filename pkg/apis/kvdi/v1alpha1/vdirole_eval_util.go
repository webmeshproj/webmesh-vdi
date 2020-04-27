package v1alpha1

import "regexp"

type APIAction struct {
	Verb              Verb     `json:"verb"`
	ResourceType      Resource `json:"resourceType"`
	ResourceName      string   `json:"resourceName"`
	ResourceNamespace string   `json:"resourceNamespace,omitempty"`
}

func (r *VDIRole) Evaluate(action *APIAction) bool {
	for _, rule := range r.Rules {
		if ok := rule.Evaluate(action); ok {
			return true
		}
	}
	return false
}

func (r *Rule) Evaluate(action *APIAction) bool {
	if !r.HasVerb(action.Verb) {
		return false
	}
	if !r.HasResourceType(action.ResourceType) {
		return false
	}
	if !r.MatchesResourceName(action.ResourceName) {
		return false
	}
	if action.ResourceNamespace != "" {
		if !r.HasNamespace(action.ResourceNamespace) {
			return false
		}
	}
	return true
}

func (r *Rule) HasVerb(verb Verb) bool {
	for _, item := range r.Verbs {
		if item == VerbAll {
			return true
		}
		if item == verb {
			return true
		}
	}
	return false
}

func (r *Rule) HasResourceType(resource Resource) bool {
	for _, item := range r.Resources {
		if item == ResourceAll {
			return true
		}
		if item == resource {
			return true
		}
	}
	return false
}

func (r *Rule) MatchesResourceName(name string) bool {
	for _, pattern := range r.ResourcePatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			// Should have an external validator to let the user know
			// there is a bad regex.
			continue
		}
		if re.MatchString(name) {
			return true
		}
	}
	return false
}

func (r *Rule) HasNamespace(ns string) bool {
	for _, item := range r.Namespaces {
		if item == NamespaceAll {
			return true
		}
		if item == ns {
			return true
		}
	}
	return false
}
