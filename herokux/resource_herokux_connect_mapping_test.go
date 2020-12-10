package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

const (
	TestConnectMappingBasic = `
{
    "mappings": [
        {
            "object_name": "AcceptedEventRelation",
            "config": {
                "access": "read_only",
                "sf_notify_enabled": false,
                "sf_polling_seconds": 600,
                "sf_max_daily_api_calls": 30000,
                "fields": {
                    "CreatedDate": {},
                    "Id": {},
                    "IsDeleted": {},
                    "SystemModstamp": {}
                },
                "indexes": {
                    "Id": {
                        "unique": true
                    },
                    "SystemModstamp": {
                        "unique": false
                    }
                }
            }
        }
    ],
    "version": 1
}
`
	TestConnectMappingMore = `
{
    "mappings": [
		{
            "object_name": "Account",
            "config": {
                "access": "read_only",
                "sf_notify_enabled": false,
                "sf_polling_seconds": 600,
                "sf_max_daily_api_calls": 30000,
                "fields": {
                    "CreatedDate": {},
                    "Id": {},
                    "IsDeleted": {},
                    "Name": {},
                    "SystemModstamp": {}
                },
                "indexes": {
                    "Id": {
                        "unique": true
                    },
                    "SystemModstamp": {
                        "unique": false
                    }
                }
            }
        },
        {
            "object_name": "AccountShare",
            "config": {
                "access": "read_only",
                "sf_notify_enabled": false,
                "sf_polling_seconds": 600,
                "sf_max_daily_api_calls": 30000,
                "fields": {
                    "Id": {},
                    "IsDeleted": {},
                    "LastModifiedDate": {}
                },
                "indexes": {
                    "Id": {
                        "unique": true
                    },
                    "LastModifiedDate": {
                        "unique": false
                    }
                }
            }
        }
    ],
    "version": 1
}
`
	TestConnectMappingCompressedString     = "{\"mappings\":[{\"config\":{\"access\":\"read_only\",\"fields\":{\"CreatedDate\":{},\"Id\":{},\"IsDeleted\":{},\"SystemModstamp\":{}},\"indexes\":{\"Id\":{\"unique\":true},\"SystemModstamp\":{\"unique\":false}},\"sf_max_daily_api_calls\":30000,\"sf_notify_enabled\":false,\"sf_polling_seconds\":600},\"object_name\":\"AcceptedEventRelation\"}],\"version\":1}"
	TestConnectMappingMoreCompressedString = "{\"mappings\":[{\"config\":{\"access\":\"read_only\",\"fields\":{\"CreatedDate\":{},\"Id\":{},\"IsDeleted\":{},\"Name\":{},\"SystemModstamp\":{}},\"indexes\":{\"Id\":{\"unique\":true},\"SystemModstamp\":{\"unique\":false}},\"sf_max_daily_api_calls\":30000,\"sf_notify_enabled\":false,\"sf_polling_seconds\":600},\"object_name\":\"Account\"},{\"config\":{\"access\":\"read_only\",\"fields\":{\"Id\":{},\"IsDeleted\":{},\"LastModifiedDate\":{}},\"indexes\":{\"Id\":{\"unique\":true},\"LastModifiedDate\":{\"unique\":false}},\"sf_max_daily_api_calls\":30000,\"sf_notify_enabled\":false,\"sf_polling_seconds\":600},\"object_name\":\"AccountShare\"}],\"version\":1}"
)

func TestAccHerokuxConnectMapping_Basic(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)
	connectID := testAccConfig.GetConnectIDorSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxConnectMapping_basic(appID, connectID, TestConnectMappingBasic),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_connect_mapping.foobar", "app_id", appID),
					resource.TestCheckResourceAttr(
						"herokux_connect_mapping.foobar", "connect_id", connectID),
					resource.TestCheckResourceAttr(
						"herokux_connect_mapping.foobar", "mappings", TestConnectMappingCompressedString),
					resource.TestCheckResourceAttr(
						"herokux_connect_mapping.foobar", "mapping_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"herokux_connect_mapping.foobar", "mapping_object_names.#", "1"),
					resource.TestCheckResourceAttr(
						"herokux_connect_mapping.foobar", "mapping_data.%", "1"),
				),
			},
		},
	})
}

func TestAccHerokuxConnectMapping_ExternalFileBasic(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)
	connectID := testAccConfig.GetConnectIDorSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxConnectMapping_externalFile(appID, connectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_connect_mapping.foobar", "app_id", appID),
					resource.TestCheckResourceAttr(
						"herokux_connect_mapping.foobar", "connect_id", connectID),
					resource.TestCheckResourceAttr(
						"herokux_connect_mapping.foobar", "mappings", TestConnectMappingCompressedString),
					resource.TestCheckResourceAttr(
						"herokux_connect_mapping.foobar", "mapping_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"herokux_connect_mapping.foobar", "mapping_object_names.#", "1"),
					resource.TestCheckResourceAttr(
						"herokux_connect_mapping.foobar", "mapping_data.%", "1"),
				),
			},
		},
	})
}

func TestAccHerokuxConnectMapping_Update(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)
	connectID := testAccConfig.GetConnectIDorSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxConnectMapping_basic(appID, connectID, TestConnectMappingBasic),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_connect_mapping.foobar", "app_id", appID),
					resource.TestCheckResourceAttr(
						"herokux_connect_mapping.foobar", "connect_id", connectID),
					resource.TestCheckResourceAttr(
						"herokux_connect_mapping.foobar", "mappings", TestConnectMappingCompressedString),
					resource.TestCheckResourceAttr(
						"herokux_connect_mapping.foobar", "mapping_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"herokux_connect_mapping.foobar", "mapping_object_names.#", "1"),
					resource.TestCheckResourceAttr(
						"herokux_connect_mapping.foobar", "mapping_data.%", "1"),
				),
			},
			{
				Config: testAccCheckHerokuxConnectMapping_basic(appID, connectID, TestConnectMappingMore),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_connect_mapping.foobar", "app_id", appID),
					resource.TestCheckResourceAttr(
						"herokux_connect_mapping.foobar", "connect_id", connectID),
					resource.TestCheckResourceAttr(
						"herokux_connect_mapping.foobar", "mappings", TestConnectMappingMoreCompressedString),
					resource.TestCheckResourceAttr(
						"herokux_connect_mapping.foobar", "mapping_ids.#", "2"),
					resource.TestCheckResourceAttr(
						"herokux_connect_mapping.foobar", "mapping_object_names.#", "2"),
					resource.TestCheckResourceAttr(
						"herokux_connect_mapping.foobar", "mapping_data.%", "2"),
				),
			},
		},
	})
}

func testAccCheckHerokuxConnectMapping_basic(appID, connectID, mappings string) string {
	return fmt.Sprintf(`
resource "herokux_connect_mapping" "foobar" {
	app_id = "%s"
	connect_id = "%s"
	mappings = <<-EOF
%s
EOF
}
`, appID, connectID, mappings)
}

func testAccCheckHerokuxConnectMapping_externalFile(appID, connectID string) string {
	return fmt.Sprintf(`
resource "herokux_connect_mapping" "foobar" {
	app_id = "%s"
	connect_id = "%s"
	mappings = file("test-fixtures/connect_mappings.json")
}
`, appID, connectID)
}
