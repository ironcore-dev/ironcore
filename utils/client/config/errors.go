// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package config

import "errors"

var ErrConfigNotFound = errors.New("config not found")

func IgnoreErrConfigNotFound(err error) error {
	if errors.Is(err, ErrConfigNotFound) {
		return nil
	}
	return err
}
