package sanitize

func shouldSkipSanitization(fieldName string, exemptionMap map[string]struct{}) bool {
	if exemptionMap != nil {
		if _, exempt := exemptionMap[fieldName]; exempt {
			return true
		}
	}

	// checking for config-level exemptions
	if _, exempt := configLevelExemptedFields[fieldName]; exempt {
		return true
	}

	return false
}
