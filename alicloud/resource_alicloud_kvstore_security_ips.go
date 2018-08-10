package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/r-kvstore"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAlicloudKVStoreSecurityIPs() *schema.Resource {
	return &schema.Resource{
		Create: resourceAlicloudKVStoreSecurityIPsCreate,
		Read:   resourceAlicloudKVStoreSecurityIPsRead,
		Update: resourceAlicloudKVStoreSecurityIPsUpdate,
		Delete: resourceAlicloudKVStoreSecurityIPsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"instance_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"security_ips": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
			},
			"group_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAlicloudKVStoreSecurityIPsCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AliyunClient)
	conn := client.rkvconn

	request := r_kvstore.CreateModifySecurityIpsRequest()
	request.InstanceId = d.Get("instance_id").(string)
	request.SecurityIpGroupName = d.Get("group_name").(string)
	request.SecurityIps = LOCAL_HOST_IP

	if len(d.Get("security_ips").(*schema.Set).List()) > 0 {
		request.SecurityIps = strings.Join(expandStringList(d.Get("security_ips").(*schema.Set).List())[:], COMMA_SEPARATED)
	}

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.ModifySecurityIps(request); err != nil {
			return resource.NonRetryableError(fmt.Errorf("Create security whitelist ips got an error: %#v", err))
		}
		return nil
	})

	if err != nil {
		return err
	}

	// A security ip whitelist does not have a native IP.
	d.SetId(fmt.Sprintf("%s%s%s", request.InstanceId, COLON_SEPARATED, request.SecurityIpGroupName))

	return resourceAlicloudKVStoreSecurityIPsRead(d, meta)
}

func resourceAlicloudKVStoreSecurityIPsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AliyunClient)
	conn := client.rkvconn
	instanceId := strings.Split(d.Id(), COLON_SEPARATED)[0]
	secGroupName := strings.Split(d.Id(), COLON_SEPARATED)[1]

	request := r_kvstore.CreateDescribeSecurityIpsRequest()
	request.InstanceId = instanceId
	attribs, err := conn.DescribeSecurityIps(request)
	if err != nil {
		if IsExceptedError(err, InvalidKVStoreInstanceIdNotFound) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error Describe KVStore Security IPs: %#v", err)
	}

	if attribs == nil || len(attribs.SecurityIpGroups.SecurityIpGroup) == 0 {
		d.SetId("")
		return nil
	}

	for _, secGroup := range attribs.SecurityIpGroups.SecurityIpGroup {
		if secGroup.SecurityIpGroupName == secGroupName {
			d.Set("instance_id", instanceId)
			d.Set("group_name", secGroup.SecurityIpGroupName)
			d.Set("security_ips", strings.Split(secGroup.SecurityIpList, COMMA_SEPARATED))
			return nil
		}
	}

	// No specified security group name
	d.SetId("")
	return nil
}

func resourceAlicloudKVStoreSecurityIPsUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AliyunClient)
	conn := client.rkvconn
	instanceId := strings.Split(d.Id(), COLON_SEPARATED)[0]

	if d.HasChange("group_name") || d.HasChange("security_ips") {
		ipstr := strings.Join(expandStringList(d.Get("security_ips").(*schema.Set).List())[:], COMMA_SEPARATED)
		if ipstr == "" {
			ipstr = LOCAL_HOST_IP
		}

		request := r_kvstore.CreateModifySecurityIpsRequest()
		request.InstanceId = instanceId
		request.SecurityIps = ipstr
		request.SecurityIpGroupName = d.Get("group_name").(string)
		if _, err := conn.ModifySecurityIps(request); err != nil {
			return err
		}
	}

	return resourceAlicloudKVStoreSecurityIPsRead(d, meta)
}

func resourceAlicloudKVStoreSecurityIPsDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AliyunClient)
	conn := client.rkvconn
	instanceId := strings.Split(d.Id(), COLON_SEPARATED)[0]
	ipstr := strings.Join(expandStringList(d.Get("security_ips").(*schema.Set).List())[:], COMMA_SEPARATED)
	if ipstr == "" {
		ipstr = LOCAL_HOST_IP
	}

	request := r_kvstore.CreateModifySecurityIpsRequest()
	request.InstanceId = instanceId
	request.SecurityIpGroupName = d.Get("group_name").(string)
	request.ModifyMode = "Delete"
	request.SecurityIps = ipstr
	if _, err := conn.ModifySecurityIps(request); err != nil {
		return err
	}
	return nil
}
