package bootstrap

import (
	"fmt"
	"gitee.com/DXTeam/idea-go.git/helper/config"
	sts "github.com/tencentyun/qcloud-cos-sts-sdk/go"
)

type StsClient struct {
	AppId           string
	Client          *sts.Client
	Region          string
	BucketName      string
	DurationSeconds int
}

var StsClientList = make(map[string]StsClient)

func InitSts() {
	initDefault()
}

func initDefault() {
	path := fmt.Sprintf("%sconfig/%s/tencentyun.yml", ProjectPath(), DevEnv)

	stsConfigs, err := config.GetConfig(path)

	configList, err := stsConfigs.Map("tencentyun")
	if err == nil {
		for bucketName, _ := range configList {
			secretId, _ := stsConfigs.String("tencentyun." + bucketName + ".secretId")
			secretKey, _ := stsConfigs.String("tencentyun." + bucketName + ".secretKey")
			appId, _ := stsConfigs.String("tencentyun." + bucketName + ".appId")
			region, _ := stsConfigs.String("tencentyun." + bucketName + ".Region")
			durationSeconds, _ := stsConfigs.Int("tencentyun." + bucketName + ".durationSeconds")

			c := sts.NewClient(secretId, secretKey, nil)
			StsClientList[bucketName] = StsClient{
				AppId:           appId,
				Client:          c,
				Region:          region,
				DurationSeconds: durationSeconds,
				BucketName:      fmt.Sprintf("%s-%s", bucketName, appId),
			}
		}
	}

}
