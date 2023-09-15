package checkers

import (
	"errors"
	"fmt"
	"strconv"
)

func SetMinKsta(ksta int) func(checker *NeedrestartChecker) error {
	return func(checker *NeedrestartChecker) error {
		if ksta < 1 || ksta > 2 {
			return errors.New("ksta needs to be [1, 2]")
		}

		checker.rebootMinKsta = ksta
		return nil
	}
}

func SetRebootOnSvc(rebootOnSvc bool) func(checker *NeedrestartChecker) error {
	return func(checker *NeedrestartChecker) error {
		checker.rebootOnSvc = rebootOnSvc
		return nil
	}
}

func NeedrestartCheckerFromMap(args map[string]any) (*NeedrestartChecker, error) {
	if args == nil {
		return NewNeedrestartChecker()
	}

	var opts []func(checker *NeedrestartChecker) error
	rebootOnSvcStr, ok := args["reboot_on_svc"].(string)
	if ok {
		value, err := strconv.ParseBool(rebootOnSvcStr)
		if err != nil {
			return nil, fmt.Errorf("could not parse 'reboot_on_svc' field: %w", err)
		}

		opts = append(opts, SetRebootOnSvc(value))
	}

	minKstaVal, ok := args["min_ksta"].(string)
	if ok {
		value, err := strconv.Atoi(minKstaVal)
		if err != nil {
			return nil, fmt.Errorf("could not parse 'min_ksta' field: %w", err)
		}

		opts = append(opts, SetMinKsta(value))
	}

	return NewNeedrestartChecker(opts...)
}
