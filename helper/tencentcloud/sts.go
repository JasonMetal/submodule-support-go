package tencentcloud

import (
	sts "github.com/tencentyun/qcloud-cos-sts-sdk/go"
	"idea-go/app/entity"
	"idea-go/bootstrap"
)

// GetToken 获取STS临时授权
func GetToken(bucketName, objectName string) (entity.StsToken, error) {
	token := entity.StsToken{}
	if c, ok := bootstrap.StsClientList[bucketName]; ok {
		opt := &sts.CredentialOptions{
			DurationSeconds: int64(c.DurationSeconds),
			Region:          c.Region,
			Policy:          setPolicy(c.BucketName, objectName, c.AppId),
		}
		res, err := c.Client.GetCredential(opt)
		if err != nil {
			bootstrap.CheckError(err)

			return token, err
		}

		token = entity.StsToken{
			SecretID:     res.Credentials.TmpSecretID,
			SecretKey:    res.Credentials.TmpSecretKey,
			SessionToken: res.Credentials.SessionToken,
		}
	}

	return token, nil
}

// setPolicy 设置STS权限策略
func setPolicy(bucket, objectName, appId string) *sts.CredentialPolicy {
	return &sts.CredentialPolicy{
		Statement: []sts.CredentialPolicyStatement{
			{
				// 密钥的权限列表。简单上传和分片需要以下的权限，其他权限列表请看 https://cloud.tencent.com/document/product/436/31923
				Action: getAllowAction(),
				Effect: "allow",
				Resource: []string{
					//这里改成允许的路径前缀，可以根据自己网站的用户登录态判断允许上传的具体路径，例子： a.jpg 或者 a/* 或者 * (使用通配符*存在重大安全风险, 请谨慎评估使用)
					"qcs::cos:ap-guangzhou:uid/1306132645:" + bucket + "-" + appId + "/" + objectName,
				},
			},
		},
	}
}

// getAllowAction 设置STS允许操作的行为
func getAllowAction() []string {
	return []string{
		// 简单上传
		"name/cos:PostObject",
		"name/cos:PutObject",
		// 分片上传
		"name/cos:InitiateMultipartUpload",
		"name/cos:ListMultipartUploads",
		"name/cos:ListParts",
		"name/cos:UploadPart",
		"name/cos:CompleteMultipartUpload",
	}
}
