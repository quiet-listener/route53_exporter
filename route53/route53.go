package route53

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

type Route53Data struct {
	Name         string
	Hostedzoneid string
	PrivateZone  string
	Count        float64
	Limit        float64
}

func getAwsSession() (*session.Session, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func getAwsCreds(sess *session.Session) *credentials.Credentials {
	role_arn := os.Getenv("AWS_ROLE_ARN")
	token_file := os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE")
	return stscreds.NewWebIdentityCredentials(sess, role_arn, "route53-session", token_file)
}

func getAwsAuth() (*route53.Route53, error) {
	sess, err := getAwsSession()
	if err != nil {
		return nil, err
	}
	creds := getAwsCreds(sess)
	// sess := session.Must(session.NewSessionWithOptions(session.Options{
	// 	SharedConfigState: session.SharedConfigEnable,
	// }))
	svc := route53.New(sess, &aws.Config{Credentials: creds}, aws.NewConfig())
	return svc, nil
}

func Route53Metrics() ([]*Route53Data, error) {
	HostedZoneLimitTypeMaxRrsetsByZone := "MAX_RRSETS_BY_ZONE"
	routeSess, err := getAwsAuth()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	hostedZoneList := make([]*route53.HostedZone, 0, 0)
	route53data := make([]*Route53Data, 0, 0)
	input := &route53.ListHostedZonesInput{}
	for {
		listHostedZones, err := routeSess.ListHostedZones(input)
		if err != nil {
			fmt.Println(err)
		} else {
			hostedZoneList = append(hostedZoneList, listHostedZones.HostedZones...)
		}
		if !*listHostedZones.IsTruncated {
			break
		}
		input.SetMarker(*listHostedZones.Marker)
	}
	for _, hostedzone := range hostedZoneList {

		limit, err := routeSess.GetHostedZoneLimit(&route53.GetHostedZoneLimitInput{
			HostedZoneId: hostedzone.Id,
			Type:         &HostedZoneLimitTypeMaxRrsetsByZone,
		})
		if err != nil {
			fmt.Println(err)
		}
		route53data = append(route53data, &Route53Data{
			Name:         *hostedzone.Name,
			Hostedzoneid: *hostedzone.Id,
			PrivateZone:  strconv.FormatBool(*hostedzone.Config.PrivateZone),
			Count:        float64(*hostedzone.ResourceRecordSetCount),
			Limit:        float64(*limit.Limit.Value),
		})
	}
	return route53data, nil
}
