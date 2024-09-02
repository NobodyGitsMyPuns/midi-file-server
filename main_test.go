package main

const (
	Acct = "112168818644504200034"
)

//todo set timeout on test workflow and lint workflow so it doesn't run forever if something goes wrong

// func TestInitGCP(t *testing.T) {
// 	_, err := InitGCPWithServiceAccount(GCP_project, "/Users/jesselopez/Documents/repos/midi-file-server/gothic_key.json")
// 	require.NoError(t, err)
// }
// Route to handle file listing
//    server.on("/files", HTTP_GET, []() {
// 	String fileList = "";
// 	File root = LittleFS.open("/");
// 	File file = root.openNextFile();
// 	while (file) {
// 		fileList += String(file.name()) + "\n";
// 		file = root.openNextFile();
// 	}
// 	server.send(200, "text/plain", fileList);
// });

// // Route to handle file deletion
// server.on("/delete", HTTP_DELETE, []() {
// 	if (server.hasArg("name")) {
// 		String filename = "/" + server.arg("name");
// 		if (LittleFS.remove(filename)) {
// 			Serial.println("File deleted successfully");

// 			// Send a response
// 			String response = "File Deleted";
// 			server.send(200, "text/plain", response);
// 		} else {
// 			server.send(404, "text/plain", "File Not Found");
// 		}
// 	} else {
// 		server.send(400, "text/plain", "Name parameter missing");
// 	}
// });

// // mongoDB
// // //Connect to MongoDB
// func connectMongoDBLocal() (*mongo.Client, error) {
// 	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
// 	client, err := mongo.Connect(context.Background(), clientOptions) //todod old

// 	if err != nil {
// 		return nil, err
// 	}
// 	return client, nil
// }

// // Register a user
// func RegisterUser(client *mongo.Client, serialNumber, username, password string) error {
// 	collection := client.Database("testdb").Collection("users")
// 	user := bson.D{
// 		{Key: "serial_number", Value: serialNumber},
// 		{Key: "username", Value: username},
// 		{Key: "password", Value: password},
// 	}
// 	_, err := collection.InsertOne(context.Background(), user)
// 	return err
// }

// // Login a user
// func LoginUser(client *mongo.Client, username, password string) (bool, error) {
// 	collection := client.Database("testdb").Collection("users")
// 	filter := bson.D{{Key: "username", Value: username}, {Key: "password", Value: password}}
// 	var result bson.D
// 	err := collection.FindOne(context.Background(), filter).Decode(&result)
// 	if err == mongo.ErrNoDocuments {
// 		return false, nil
// 	}
// 	return err == nil, err
// }

// func TestRegisterUser(t *testing.T) {
// 	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

// 	mt.Run("register user", func(mt *mtest.T) {
// 		client, err := connectMongoDBLocal()
// 		if err != nil {
// 			t.Fatalf("Failed to connect to MongoDB: %v", err)
// 		}

// 		err = RegisterUser(client, "12345", "testuser", "testpass")
// 		if err != nil {
// 			t.Errorf("Failed to register user: %v", err)
// 		}
// 	})
// }

// func TestLoginUser(t *testing.T) {
// 	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

// 	mt.Run("login user", func(mt *mtest.T) {
// 		client, err := connectMongoDBLocal()
// 		if err != nil {
// 			t.Fatalf("Failed to connect to MongoDB: %v", err)
// 		}

// 		//First, register a user
// 		err = RegisterUser(client, "12345", "testuser", "testpass")
// 		if err != nil {
// 			t.Fatalf("Failed to register user: %v", err)
// 		}

// 		//Now, try to log in
// 		success, err := LoginUser(client, "testuser", "testpass")
// 		if err != nil {
// 			t.Errorf("Failed to log in user: %v", err)
// 		}

// 		if !success {
// 			t.Errorf("Expected login to succeed, but it failed")
// 		}
// 	})
// }
