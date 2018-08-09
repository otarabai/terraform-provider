package alicloud

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAlicloudRKVInstancesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAlicloudKVStoreInstancesDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlicloudDataSourceID("data.alicloud_kvstore_instances.rkvs"),
					resource.TestCheckResourceAttr("data.alicloud_kvstore_instances.rkvs", "instances.#", "1"),
					resource.TestCheckResourceAttr("data.alicloud_kvstore_instances.rkvs", "instances.0.instance_class", "redis.master.small.default"),
					resource.TestCheckResourceAttr("data.alicloud_kvstore_instances.rkvs", "instances.0.name", "testAccCheckAlicloudKVStoreInstancesDataSourceConfig"),
					resource.TestCheckResourceAttr("data.alicloud_kvstore_instances.rkvs", "instances.0.instance_type", "Redis"),
					resource.TestCheckResourceAttr("data.alicloud_kvstore_instances.rkvs", "instances.0.instance_type", "Redis"),
					resource.TestCheckResourceAttr("data.alicloud_kvstore_instances.rkvs", "instances.0.instance_charge_type", string(Postpaid)),
				),
			},
		},
	})
}

const testAccCheckAlicloudKVStoreInstancesDataSourceConfig = `
data "alicloud_rkv_instances" "rkvs" {
  name_regex = "${alicloud_kvstore_instance.rkv.instance_name}"
}

resource "alicloud_kvstore_instance" "rkv" {
	instance_name = "testAccCheckAlicloudKVStoreInstancesDataSourceConfig"
	instance_class = "redis.master.small.default"
	instance_type = "Redis"
	engine_version = "2.8"
	password = "Passw0rd"
    instance_charge_type = "PostPaid"
}
`
