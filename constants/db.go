package constants

const BucketKeyList = "KEY_LIST"

const (
	BucketSatpointToSequenceNumber                      = "SATPOINT_TO_SEQUENCE_NUMBER"
	BucketSatToSequenceNumber                           = "SAT_TO_SEQUENCE_NUMBER"
	BucketHeightToBlockHeader                           = "HEIGHT_TO_BLOCK_HEADER"
	BucketHeightToLastSequenceNumber                    = "HEIGHT_TO_LAST_SEQUENCE_NUMBER"
	BucketInscriptionIdToSequenceNumber                 = "INSCRIPTION_ID_TO_SEQUENCE_NUMBER"
	BucketInscriptionNumberToSequenceNumber             = "INSCRIPTION_NUMBER_TO_SEQUENCE_NUMBER"
	BucketOutpointToValue                               = "OUTPOINT_TO_VALUE"
	BucketOutpointToInscriptions                        = "OUTPOINT_TO_INSCRIPTIONS"
	BucketSatToSatpoint                                 = "SAT_TO_SATPOINT"
	BucketSequenceNumberToInscriptionEntry              = "SEQUENCE_NUMBER_TO_INSCRIPTION_ENTRY"
	BucketSequenceNumberToSatpoint                      = "SEQUENCE_NUMBER_TO_SATPOINT"
	BucketStatisticToCount                              = "STATISTIC_TO_COUNT"
	BucketWriteTransactionStartingBlockCountToTimestamp = "WRITE_TRANSACTION_STARTING_BLOCK_COUNT_TO_TIMESTAMP"
)

var KVBuckets = []string{
	BucketSatToSequenceNumber,
	BucketHeightToBlockHeader,
	BucketHeightToLastSequenceNumber,
	BucketInscriptionIdToSequenceNumber,
	BucketInscriptionNumberToSequenceNumber,
	BucketOutpointToValue,
	BucketOutpointToInscriptions,
	BucketSatToSatpoint,
	BucketSequenceNumberToInscriptionEntry,
	BucketSequenceNumberToSatpoint,
	BucketStatisticToCount,
	BucketWriteTransactionStartingBlockCountToTimestamp,
}

var ListBuckets = []string{
	BucketSatpointToSequenceNumber,
}

type Statistic string

const (
	StatisticSchema              Statistic = "0"
	StatisticBlessedInscriptions Statistic = "1"
	StatisticCommits             Statistic = "2"
	StatisticCursedInscriptions  Statistic = "3"
	// StatisticIndexRunes          Statistic = "4"
	// StatisticIndexSats           Statistic = "5"
	// StatisticLostSats            Statistic = "6"
	StatisticOutputsTraversed Statistic = "7"
	// StatisticReservedRunes       Statistic = "8"
	// StatisticRunes               Statistic = "9"
	// StatisticSatRanges           Statistic = "10"
	StatisticUnboundInscriptions Statistic = "11"
	StatisticIndexTransactions   Statistic = "12"
)
