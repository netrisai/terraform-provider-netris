/*
Copyright 2021. Netris, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package inventoryprofile

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
)

func validateIPPrefix(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	re := regexp.MustCompile(`(^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\/([0-9]|[12]\d|3[0-2]))?$)|(^((([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|(([0-9A-Fa-f]{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]{1,4})?:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:)))(%.+)?(\/([1-9]|[1-5][0-9]|6[0-4]))?$)`)
	if !re.Match([]byte(v)) {
		errs = append(errs, fmt.Errorf("invalid %s: %s", key, v))
	}
	return warns, errs
}

func validateIP(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if !validateIPAddr(v) {
		errs = append(errs, fmt.Errorf("invalid %s: %s", key, v))
	}
	return warns, errs
}

func validateIPAddr(s string) bool {
	re := regexp.MustCompile(`(^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\/([0-9]|[12]\d|3[0-2]))?$)|(^((([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|(([0-9A-Fa-f]{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]{1,4})?:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:)))(%.+)?(\/([0-9]|[1-5][0-9]|6[0-4]))?$)`)
	return re.Match([]byte(s))
}

func validateFQDN(s string) bool {
	re := regexp.MustCompile(`^(.{1,22}$)?(([a-z0-9-]{1,63}\.)?(xn--+)?[a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,63}$`)
	return re.Match([]byte(s))
}

func validateNTP(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if !validateIPAddr(v) && !validateFQDN(v) {
		errs = append(errs, fmt.Errorf("invalid %s: %s", key, v))
	}
	return warns, errs
}

func validateTimeZone(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if _, ok := timezones[v]; !ok {
		errs = append(errs, fmt.Errorf("invalid %s: %s", key, v))
	}
	return warns, errs
}

func validatePort(val interface{}, key string) (warns []string, errs []error) {
	if _, err := valPort(val.(string)); err != nil && val.(string) != "" {
		errs = append(errs, fmt.Errorf(`Invalid value "%s". %s`, val.(string), err))
	}
	return warns, errs
}

func valPort(port string) (int, error) {
	log.Println("[DEBUG] port", port)
	v, err := strconv.Atoi(port)
	if err != nil {
		return 0, fmt.Errorf(`Port should be a number`)
	} else if !(v >= 1 && v <= 65535) {
		return 0, fmt.Errorf("Port should be in range 1-65535")
	}
	return v, nil
}

func validateProtocol(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if !(v == "any" || v == "tcp" || v == "udp") {
		errs = append(errs, fmt.Errorf("invalid protocol. Available values are (any, tcp, udp)"))
	}
	return warns, errs
}

var timezones = map[string]int{
	"Pacific/Niue":                   1,
	"Pacific/Pago_Pago":              1,
	"Pacific/Honolulu":               1,
	"Pacific/Rarotonga":              1,
	"Pacific/Tahiti":                 1,
	"Pacific/Marquesas":              1,
	"America/Anchorage":              1,
	"Pacific/Gambier":                1,
	"America/Los_Angeles":            1,
	"America/Tijuana":                1,
	"America/Vancouver":              1,
	"America/Whitehorse":             1,
	"Pacific/Pitcairn":               1,
	"America/Denver":                 1,
	"America/Phoenix":                1,
	"America/Mazatlan":               1,
	"America/Dawson_Creek":           1,
	"America/Edmonton":               1,
	"America/Hermosillo":             1,
	"America/Yellowknife":            1,
	"America/Belize":                 1,
	"America/Chicago":                1,
	"America/Mexico_City":            1,
	"America/Regina":                 1,
	"America/Tegucigalpa":            1,
	"America/Winnipeg":               1,
	"America/Costa_Rica":             1,
	"America/El_Salvador":            1,
	"Pacific/Galapagos":              1,
	"America/Guatemala":              1,
	"America/Managua":                1,
	"America/Cancun":                 1,
	"America/Bogota":                 1,
	"Pacific/Easter":                 1,
	"America/New_York":               1,
	"America/Iqaluit":                1,
	"America/Toronto":                1,
	"America/Guayaquil":              1,
	"America/Havana":                 1,
	"America/Jamaica":                1,
	"America/Lima":                   1,
	"America/Nassau":                 1,
	"America/Panama":                 1,
	"America/Port-au-Prince":         1,
	"America/Rio_Branco":             1,
	"America/Halifax":                1,
	"America/Barbados":               1,
	"Atlantic/Bermuda":               1,
	"America/Boa_Vista":              1,
	"America/Caracas":                1,
	"America/Curacao":                1,
	"America/Grand_Turk":             1,
	"America/Guyana":                 1,
	"America/La_Paz":                 1,
	"America/Manaus":                 1,
	"America/Martinique":             1,
	"America/Port_of_Spain":          1,
	"America/Porto_Velho":            1,
	"America/Puerto_Rico":            1,
	"America/Santo_Domingo":          1,
	"America/Thule":                  1,
	"America/St_Johns":               1,
	"America/Araguaina":              1,
	"America/Asuncion":               1,
	"America/Belem":                  1,
	"America/Argentina/Buenos_Aires": 1,
	"America/Campo_Grande":           1,
	"America/Cayenne":                1,
	"America/Cuiaba":                 1,
	"America/Fortaleza":              1,
	"America/Godthab":                1,
	"America/Maceio":                 1,
	"America/Miquelon":               1,
	"America/Montevideo":             1,
	"Antarctica/Palmer":              1,
	"America/Paramaribo":             1,
	"America/Punta_Arenas":           1,
	"America/Recife":                 1,
	"Antarctica/Rothera":             1,
	"America/Bahia":                  1,
	"America/Santiago":               1,
	"Atlantic/Stanley":               1,
	"America/Noronha":                1,
	"America/Sao_Paulo":              1,
	"Atlantic/South_Georgia":         1,
	"Atlantic/Azores":                1,
	"Atlantic/Cape_Verde":            1,
	"America/Scoresbysund":           1,
	"Africa/Abidjan":                 1,
	"Africa/Accra":                   1,
	"Africa/Bissau":                  1,
	"Atlantic/Canary":                1,
	"Africa/Casablanca":              1,
	"America/Danmarkshavn":           1,
	"Europe/Dublin":                  1,
	"Africa/El_Aaiun":                1,
	"Atlantic/Faroe":                 1,
	"Etc/GMT":                        1,
	"Europe/Lisbon":                  1,
	"Europe/London":                  1,
	"Africa/Monrovia":                1,
	"Atlantic/Reykjavik":             1,
	"Africa/Algiers":                 1,
	"Europe/Amsterdam":               1,
	"Europe/Andorra":                 1,
	"Europe/Berlin":                  1,
	"Europe/Brussels":                1,
	"Europe/Budapest":                1,
	"Europe/Belgrade":                1,
	"Europe/Prague":                  1,
	"Africa/Ceuta":                   1,
	"Europe/Copenhagen":              1,
	"Europe/Gibraltar":               1,
	"Africa/Lagos":                   1,
	"Europe/Luxembourg":              1,
	"Europe/Madrid":                  1,
	"Europe/Malta":                   1,
	"Europe/Monaco":                  1,
	"Africa/Ndjamena":                1,
	"Europe/Oslo":                    1,
	"Europe/Paris":                   1,
	"Europe/Rome":                    1,
	"Europe/Stockholm":               1,
	"Europe/Tirane":                  1,
	"Africa/Tunis":                   1,
	"Europe/Vienna":                  1,
	"Europe/Warsaw":                  1,
	"Europe/Zurich":                  1,
	"Asia/Amman":                     1,
	"Europe/Athens":                  1,
	"Asia/Beirut":                    1,
	"Europe/Bucharest":               1,
	"Africa/Cairo":                   1,
	"Europe/Chisinau":                1,
	"Asia/Damascus":                  1,
	"Asia/Gaza":                      1,
	"Europe/Helsinki":                1,
	"Asia/Jerusalem":                 1,
	"Africa/Johannesburg":            1,
	"Africa/Khartoum":                1,
	"Europe/Kiev":                    1,
	"Africa/Maputo":                  1,
	"Europe/Kaliningrad":             1,
	"Asia/Nicosia":                   1,
	"Europe/Riga":                    1,
	"Europe/Sofia":                   1,
	"Europe/Tallinn":                 1,
	"Africa/Tripoli":                 1,
	"Europe/Vilnius":                 1,
	"Africa/Windhoek":                1,
	"Asia/Baghdad":                   1,
	"Europe/Istanbul":                1,
	"Europe/Minsk":                   1,
	"Europe/Moscow":                  1,
	"Africa/Nairobi":                 1,
	"Asia/Qatar":                     1,
	"Asia/Riyadh":                    1,
	"Antarctica/Syowa":               1,
	"Asia/Tehran":                    1,
	"Asia/Baku":                      1,
	"Asia/Dubai":                     1,
	"Indian/Mahe":                    1,
	"Indian/Mauritius":               1,
	"Europe/Samara":                  1,
	"Indian/Reunion":                 1,
	"Asia/Tbilisi":                   1,
	"Asia/Yerevan":                   1,
	"Asia/Kabul":                     1,
	"Asia/Aqtau":                     1,
	"Asia/Aqtobe":                    1,
	"Asia/Ashgabat":                  1,
	"Asia/Dushanbe":                  1,
	"Asia/Karachi":                   1,
	"Indian/Kerguelen":               1,
	"Indian/Maldives":                1,
	"Antarctica/Mawson":              1,
	"Asia/Yekaterinburg":             1,
	"Asia/Tashkent":                  1,
	"Asia/Colombo":                   1,
	"Asia/Kolkata":                   1,
	"Asia/Kathmandu":                 1,
	"Asia/Almaty":                    1,
	"Asia/Bishkek":                   1,
	"Indian/Chagos":                  1,
	"Asia/Dhaka":                     1,
	"Asia/Omsk":                      1,
	"Asia/Thimphu":                   1,
	"Antarctica/Vostok":              1,
	"Indian/Cocos":                   1,
	"Asia/Yangon":                    1,
	"Asia/Bangkok":                   1,
	"Indian/Christmas":               1,
	"Antarctica/Davis":               1,
	"Asia/Saigon":                    1,
	"Asia/Hovd":                      1,
	"Asia/Jakarta":                   1,
	"Asia/Krasnoyarsk":               1,
	"Asia/Brunei":                    1,
	"Asia/Shanghai":                  1,
	"Asia/Choibalsan":                1,
	"Asia/Hong_Kong":                 1,
	"Asia/Kuala_Lumpur":              1,
	"Asia/Macau":                     1,
	"Asia/Makassar":                  1,
	"Asia/Manila":                    1,
	"Asia/Irkutsk":                   1,
	"Asia/Singapore":                 1,
	"Asia/Taipei":                    1,
	"Asia/Ulaanbaatar":               1,
	"Australia/Perth":                1,
	"Asia/Pyongyang":                 1,
	"Asia/Dili":                      1,
	"Asia/Jayapura":                  1,
	"Asia/Yakutsk":                   1,
	"Pacific/Palau":                  1,
	"Asia/Seoul":                     1,
	"Asia/Tokyo":                     1,
	"Australia/Darwin":               1,
	"Antarctica/DumontDUrville":      1,
	"Australia/Brisbane":             1,
	"Pacific/Guam":                   1,
	"Asia/Vladivostok":               1,
	"Pacific/Port_Moresby":           1,
	"Pacific/Chuuk":                  1,
	"Australia/Adelaide":             1,
	"Antarctica/Casey":               1,
	"Australia/Hobart":               1,
	"Australia/Sydney":               1,
	"Pacific/Efate":                  1,
	"Pacific/Guadalcanal":            1,
	"Pacific/Kosrae":                 1,
	"Asia/Magadan":                   1,
	"Pacific/Norfolk":                1,
	"Pacific/Noumea":                 1,
	"Pacific/Pohnpei":                1,
	"Pacific/Funafuti":               1,
	"Pacific/Kwajalein":              1,
	"Pacific/Majuro":                 1,
	"Asia/Kamchatka":                 1,
	"Pacific/Nauru":                  1,
	"Pacific/Tarawa":                 1,
	"Pacific/Wake":                   1,
	"Pacific/Wallis":                 1,
	"Pacific/Auckland":               1,
	"Pacific/Enderbury":              1,
	"Pacific/Fakaofo":                1,
	"Pacific/Fiji":                   1,
	"Pacific/Tongatapu":              1,
	"Pacific/Apia":                   1,
	"Pacific/Kiritimati":             1,
}
