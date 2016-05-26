#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#include <linux/i2c-dev.h>
#include <sys/ioctl.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>

char device_address;

#define BUFSIZE 4

char value, phase, phase_write, number;

/* Suche nach I2C-Adressen */
void scan_i2c_bus(int device)
  {
  int port, res;

  for (port = 0; port < 127; port++)
  {
      if (ioctl(device, I2C_SLAVE, port) < 0)
      //if (ioctl(device, 0x70, port) < 0)
      perror("ioctl() I2C_SLAVE failed\n");
      else
      {
      res = i2c_smbus_read_byte(device);
      if (res >= 0)
      {
        //printf("i2c chip found at: %x, val = %d\n", port, res);
	 device_address=port;
      }

      }
    }
  }





//int main(void)
int main( int argc, char* argv[] )
{
  int device;
  unsigned long funcs;
  char Daten[BUFSIZE];
  int i,t;
  int z;

  float ergebnis;  
  float offset_current;  
  float effektivwert;
  float summe_effektivwert;
  float rms_faktor_current=4191910;
  float strom_klemme;
  float strom_real;
  float offset_voltage;   
  float rms_faktor_voltage=2427873; //Hier nicht 4191910, sondern 2427873 (vorab ermittelt)
  float summe_ergebnis; 
  float spannung_real;  
  float power;
  float faktor_1=256;
  float faktor_2=65536;
  float faktor_3=16777216;
 

value=atoi(argv[2]);
//printf("\r\nEingabe Value %i\r\n",value);
phase=atoi(argv[3]);
//printf("\r\nEingabe Phase %i\r\n",phase);
		

/*		
//Leistungsmessung!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
//AN ALLEN DREI EINGÄNGEN...
*/

    offset_current=0.97129167;//OFFSET HABE ICH HIER MAL BEI 5A ermittelt!!!´
    offset_voltage=1.00;
    



  //printf("Opening device...");
  if ((device = open("/dev/i2c-0", O_RDWR)) < 0)//HABE ICH GEÄNDERT VON /dev/i2c-1 auf /dev/i2c-0
  {
    //perror("open() failed");
    exit (1);
  }
  

  if (ioctl(device,I2C_FUNCS,&funcs) < 0)
    {
    	//perror("ioctl() I2C_FUNCS failed");
    	exit (1);
    }

  /*if (funcs & I2C_FUNC_I2C)
    printf("I2C\n");
  if (funcs & (I2C_FUNC_SMBUS_BYTE))
    printf("I2C_FUNC_SMBUS_BYTE\n");
  */


  //GANZ WICHTIG!!!
  //WENN PSM0-MODE NICHT VORHER AKTIVIERT, PASSIERT GAR NICHTS!!!

  //ZUR AKTIVIERUNG: PM0-PIN auf HIGH und PM1-PIN auf LOW!!!!!
  //Nach dem Einschalten des Raspberrys wird der Raspberry erstmal eine Fehlermeldung melden,
  //wenn die PINS auf HIGH, bzw. LOW gesetzt werden, das ist normal. Wenn die PINS das zweite mal gesetzt werden, wird der Befehl auch umgesetzt!!!
  system ("echo 23 > /sys/class/gpio/export");//GPIO 23 (PIN 16) initialisieren!!!
  system ("echo 24 > /sys/class/gpio/export");//GPIO 24 (PIN 18) initialisieren!!!
  system ("echo high > /sys/class/gpio/gpio23/direction");//PSM0 (HIGH)
  system ("echo low > /sys/class/gpio/gpio24/direction");//PSM1 (LOW)

    scan_i2c_bus(device);//Herausfinden der Device-Addresse!!!
	
    device_address=0x38;//Die Adresse des ADE7878 ist 0x38. Am Ende dieser Adresse kommt noch ein Bit: 1 für Read und 0 für write, deswegen 0x71, bzw. 0x70.

    if (ioctl(device, I2C_SLAVE, device_address) < 0)
    //if (ioctl(device, 0x70, device_address) < 0)
    {
       //printf("\r\nDeviceadresse wurde nicht gefunden!\n");
       exit(1);
    }
    /*else
    {
	printf("\r\nDeviceadresse wurde gefunden!\n");
	printf("\r\nI2C_SLAVE: %x\n",I2C_SLAVE);
	//printf("\r\nI2C_SLAVE: %x\n",0x70);
	printf("\r\nDEVICE-ADDRESS: %x\n",device_address);
    }*/


//CONFIG2-REGISTER (I2C-LOCK) WRITE!!!!
//-------------------------------------------------------------------------------------

	//Über das CONFIG2-REGISTER speichere ich die gewählte Datenschnittstelle, also bei uns "I2C"!!!
	//Dieses Register verhindert, dass im laufenden Betrieb aus versehen auf "SPI" umgestellt wird!!
	


	Daten[0] = 0xEC;//0xEC01 (CONFIG2-REGISTER)
   	Daten[1] = 0x01; 
 	Daten[2] = 0x02;//00000010 --> Bedeutet I2C-Lock (I2C ist nun die gewählte Übertragungsart)
	
	

 	if(write(device, Daten, 3) != 3)//Anzahl der zu übertragenen Bytes hier einsetzen (hier 3 --> Register plus Dateninhalt)

    	{
    		//printf("Fehler beim Schreiben der Daten (I2C-LOCK)!\n");
    		return -1;
    	}
//-------------------------------------------------------------------------------------


	
//CONFIG2-REGISTER (I2C-LOCK) READ!!!!
//-------------------------------------------------------------------------------------
	

	//Hier wird überprüft, ob das CONFIG2-REGISTER (I2C-LOCK)richtig gesetzt wurde (Wird allerdings derzeit nicht ausgewertet)!!!!
	Daten[0] = 0xEC;//0xEC01 (CONFIG2-REGISTER)
   	Daten[1] = 0x01; 
	//Daten[2] = 0x00; 
 		
	

 	if(write(device, Daten, 1) != 1)
    	{
    		//printf("Fehler beim Schreiben der Daten (I2C-LOCK)!\n");
    		return -1;
    	}
	
	if (read(device, Daten, BUFSIZE) != BUFSIZE)
  	{
    		//printf("Fehler beim Lesen der Daten!\n");
    		return -1;
    	}
    	/*else
    	{
      		//printf("\r\nI2C-LOCK-Register:\n");

      		for(i = 0; i < 2; i++)
      		{
      		printf("%x,",Daten[i]);
      		}
    	}
	*/

//-------------------------------------------------------------------------------------


//DICOEFF-REGISTER (Digital-Integrator) WRITE!!!
//-------------------------------------------------------------------------------------

	
	Daten[0] = 0x43;//0x43B5 (DICOEFF-REGISTER)
   	Daten[1] = 0xB5; 
 	Daten[2] = 0xFF;
	Daten[3] = 0x80;
	Daten[4] = 0x00;
	
	

 	if(write(device, Daten, 5) != 5)//Anzahl der zu übertragenen Bytes hier einsetzen (hier 5 --> Register plus Dateninhalt)

    	{
    		//printf("Fehler beim Schreiben der Daten (DICOEFF-REGISTER)!\n");
    		return -1;
    	}
	
	
//-------------------------------------------------------------------------------------



//STARTE RUN  (WRITE)
//-------------------------------------------------------------------------------------

	
	//Vor jeder Messwertabfrage muss ein "RUN-Befehl", also das RUN-Register, beschrieben werden!!!

	Daten[0] = 0xE2;//0xE228 (RUN-Register)
   	Daten[1] = 0x28; 
 	Daten[2] = 0x00;
	Daten[3] = 0x01;
	//Daten[4] = 0x00;
	//Daten[5] = 0x01;
	

 	if(write(device, Daten, 4) != 4)//Anzahl der zu übertragenen Bytes hier einsetzen (hier 4 --> Register plus Dateninhalt)
    	{
    		//printf("Fehler beim Schreiben der Daten!\n");
    		return -1;
    	}
	

//-------------------------------------------------------------------------------------



//STARTE RUN  (READ)
//-------------------------------------------------------------------------------------

	
	//Hier wird überprüft, ob das RUN-Register auch gesetzt wird (Wird allerdings derzeit nicht ausgewertet)!!

	Daten[0] = 0xE2;//0xE228 (RUN-Register)
   	Daten[1] = 0x28; 
 	
	

 	if(write(device, Daten, 2) != 2)
    	{
    		//printf("Fehler beim Schreiben der Daten!\n");
    		return -1;
    	}
	
	if (read(device, Daten, BUFSIZE) != BUFSIZE)
  	{
    		//printf("Fehler beim Lesen der Daten!\n");
    		return -1;
    	}
    	
//-------------------------------------------------------------------------------------

//JSON
printf("{");
printf("\"serial\": \"12345678\",");
printf("\"time\": \"2016-03-11 16:06:00\",");
printf("\"softwareversion\": \"1.0.1\",");
printf("\"ipaddress\": \"10.1.0.22\",");
printf("\"datasets\": [{");
printf("\"time\": \"2016-03-11 11:20:00\",");
printf("\"phases\": [{");



if (phase!=77)number=3;//1 Phase
else number=0;//All
		
for (t=number;t<4;t++)
{
	if(phase==77)
	{
	   	phase_write=t+1;
	}
	else phase_write=phase;

	switch(phase_write)
	{
		
		//Phase 1
		case 1:
		{
			
			printf("\"phase\": %i,",phase_write);
			printf("\"name\": \"phase %i\",",phase_write);
			printf("\"values\": [{");
 		
			switch (value)
			{
				//Current:
				case 30:
				{
										
						summe_effektivwert=0;
						strom_klemme=0;
						ergebnis=0;
						effektivwert=0;
						strom_real=0;

 						for (z=0; z<500; z++) //In dieser Schleife werden 500 Strommessungen durchgeführt und anschließend der Mittelwert berechnet!!!
 						{

  							Daten[0] = 0x43;//0x43C0 (AIRMS; Current rms an A)
  				
							if (phase_write==1)	   Daten[1] = 0xC0;//Current Phase A
							else if (phase_write==2)Daten[1] = 0xC2;//Current Phase B
							else if (phase_write==3)Daten[1] = 0xC4;//Current Phase C
							else if (phase_write==4)Daten[1] = 0xC6;//Current N

  	 	 					if(write(device, Daten, 2) != 2)
  	 	 					{
  	 			 				//printf("Fehler beim Schreiben der Daten!\n");
  	  			 				return -1;
 	 	 					}
  	  
  	  						if (read(device, Daten, BUFSIZE) != BUFSIZE)
  	  						{
  	 	 	 					//printf("Fehler beim Lesen der Daten!\n");
 	  	  	 					return -1;
							}
 	 
  	 						ergebnis=(faktor_3*Daten[0])+(faktor_2*Daten[1])+(faktor_1*Daten[2])+Daten[3];   
       
 	 						effektivwert=(ergebnis*0.3535)/(rms_faktor_current); 	 

        						summe_effektivwert=summe_effektivwert+effektivwert;//Hier werden die 500 Messungen aufsummiert!!!
						}

  	 					effektivwert=summe_effektivwert/500; //Hier wird der Mittelwert von 500 Messungen berechnet!!!	
    
   	 					strom_klemme=effektivwert/7.07107;//10 Ohm ist der Widerstand der Stromklemme, parallel mit Widerstand (7,07107 Ohm)
  
   	 					offset_current=0.97129167;//OFFSET HABE ICH HIER MAL BEI 5A ermittelt!!!   

   	 					strom_real=(strom_klemme/0.05)*100*offset_current; //OFFSET HABE ICH HIER MAL BEI 5A ermittelt!!!

	 					//printf("\r\nStrom Phase:%i (effektiv): %f A",phase_write,strom_real);
					
						printf("\"type\": \"current\",");
						printf("\"unity\": \"A\",");
						printf("\"data\": %f",strom_real);
						printf("}, {");



						spannung_real=0;
						ergebnis=0;
						summe_ergebnis=0;

 						for (z=0; z<500; z++) //In dieser Schleife werden 500 Strommessungen durchgeführt und anschließend der Mittelwert berechnet!!!
 						{

  							Daten[0] = 0x43;//0x43C1 (AVRMS; Spannung rms an VA)
  							Daten[1] = 0xC1; 
 

  							if(write(device, Daten, 2) != 2)
  							{
  	  							//printf("Fehler beim Schreiben der Daten!\n");
  	 							return -1;
 							}

      


  	 
  	 						if (read(device, Daten, BUFSIZE) != BUFSIZE)
  	 						{
  	  							//printf("Fehler beim Lesen der Daten!\n");
 	 							return -1;
 	 						} 
	
  	 						ergebnis=(faktor_3*Daten[0])+(faktor_2*Daten[1])+(faktor_1*Daten[2])+Daten[3];    	
       

        						summe_ergebnis=summe_ergebnis+ergebnis;

							//Ich mache es hier anders als bei der Strommessung.
							//Ich habe bei 229,8 V (eff.) einen Wert von 2427873 erhalten
							//Ich summiere hier einfach 500 mal die gemessenen Werte auf und ermittle anschließend die Spannung!
						}

  	 					ergebnis=summe_ergebnis/500; //Hier wird der Mittelwert von 500 Messungen berechnet!!! 
  						//2427873 entspricht ca. 229.8 V, somit hat man bei maximalen 4191910 einen maximal zu messenen Wert von ca. 400 V (EFFEKTIV) 
   						spannung_real=(ergebnis/rms_faktor_voltage)*229.8*offset_voltage; //OFFSET HABE ICH HIER ERSTMAL AUF 1 gesetzt!!! rms_faktor ist hier 2427873 und nicht 4191910!					
						printf("\"type\": \"voltage\",");
						printf("\"unity\": \"V\",");
						printf("\"data\": %f",spannung_real);
						printf("}, {");

						//Power:
						power=strom_real*spannung_real;
						printf("\"type\": \"power\",");
						printf("\"unity\": \"W\",");
						printf("\"data\": %f",power);
						if(phase==77)printf("}]");
						if(phase==77)printf("}, {");						

					

					break;
				

		  	 	}//Ende case 30 power (phase 1)




	 		}//Ende switch(value)
		
			break;
		}//Ende Case 1 (Phase 1)













		//Phase 2
		case 2:
		{
			
			printf("\"phase\": %i,",phase_write);
			printf("\"name\": \"phase %i\",",phase_write);
			printf("\"values\": [{");
 		
			switch (value)
			{
				//Current:
				case 30:
				{
										
						summe_effektivwert=0;
						strom_klemme=0;
						ergebnis=0;
						effektivwert=0;
						strom_real=0;

 						for (z=0; z<500; z++) //In dieser Schleife werden 500 Strommessungen durchgeführt und anschließend der Mittelwert berechnet!!!
 						{

  							Daten[0] = 0x43;//0x43C0 (AIRMS; Current rms an A)
  				
							if (phase_write==1)	   Daten[1] = 0xC0;//Current Phase A
							else if (phase_write==2)Daten[1] = 0xC2;//Current Phase B
							else if (phase_write==3)Daten[1] = 0xC4;//Current Phase C
							else if (phase_write==4)Daten[1] = 0xC6;//Current N

  	 	 					if(write(device, Daten, 2) != 2)
  	 	 					{
  	 			 				//printf("Fehler beim Schreiben der Daten!\n");
  	  			 				return -1;
 	 	 					}
  	  
  	  						if (read(device, Daten, BUFSIZE) != BUFSIZE)
  	  						{
  	 	 	 					//printf("Fehler beim Lesen der Daten!\n");
 	  	  	 					return -1;
							}
 	 
  	 						ergebnis=(faktor_3*Daten[0])+(faktor_2*Daten[1])+(faktor_1*Daten[2])+Daten[3];   
       
 	 						effektivwert=(ergebnis*0.3535)/(rms_faktor_current); 	 

        						summe_effektivwert=summe_effektivwert+effektivwert;//Hier werden die 500 Messungen aufsummiert!!!
						}

  	 					effektivwert=summe_effektivwert/500; //Hier wird der Mittelwert von 500 Messungen berechnet!!!	
    
   	 					strom_klemme=effektivwert/7.07107;//10 Ohm ist der Widerstand der Stromklemme, parallel mit Widerstand (7,07107 Ohm)
  
   	 					offset_current=0.97129167;//OFFSET HABE ICH HIER MAL BEI 5A ermittelt!!!   

   	 					strom_real=(strom_klemme/0.05)*100*offset_current; //OFFSET HABE ICH HIER MAL BEI 5A ermittelt!!!

	 					//printf("\r\nStrom Phase:%i (effektiv): %f A",phase_write,strom_real);
					
						printf("\"type\": \"current\",");
						printf("\"unity\": \"A\",");
						printf("\"data\": %f",strom_real);
						printf("}, {");



						spannung_real=0;
						ergebnis=0;
						summe_ergebnis=0;

 						for (z=0; z<500; z++) //In dieser Schleife werden 500 Strommessungen durchgeführt und anschließend der Mittelwert berechnet!!!
 						{

  							Daten[0] = 0x43;//0x43C3 (BVRMS; Spannung rms an VB)
  							Daten[1] = 0xC3; 
 

  							if(write(device, Daten, 2) != 2)
  							{
  	  							//printf("Fehler beim Schreiben der Daten!\n");
  	 							return -1;
 							}

      


  	 
  	 						if (read(device, Daten, BUFSIZE) != BUFSIZE)
  	 						{
  	  							//printf("Fehler beim Lesen der Daten!\n");
 	 							return -1;
 	 						} 
	
  	 						ergebnis=(faktor_3*Daten[0])+(faktor_2*Daten[1])+(faktor_1*Daten[2])+Daten[3];    	
       

        						summe_ergebnis=summe_ergebnis+ergebnis;

							//Ich mache es hier anders als bei der Strommessung.
							//Ich habe bei 229,8 V (eff.) einen Wert von 2427873 erhalten
							//Ich summiere hier einfach 500 mal die gemessenen Werte auf und ermittle anschließend die Spannung!
						}

  	 					ergebnis=summe_ergebnis/500; //Hier wird der Mittelwert von 500 Messungen berechnet!!! 
  						//2427873 entspricht ca. 229.8 V, somit hat man bei maximalen 4191910 einen maximal zu messenen Wert von ca. 400 V (EFFEKTIV) 
   						spannung_real=(ergebnis/rms_faktor_voltage)*229.8*offset_voltage; //OFFSET HABE ICH HIER ERSTMAL AUF 1 gesetzt!!! rms_faktor ist hier 2427873 und nicht 4191910!					
						printf("\"type\": \"voltage\",");
						printf("\"unity\": \"V\",");
						printf("\"data\": %f",spannung_real);
						printf("}, {");

						//Power:
						power=strom_real*spannung_real;
						printf("\"type\": \"power\",");
						printf("\"unity\": \"W\",");
						printf("\"data\": %f",power);
						if(phase==77)printf("}]");
						if(phase==77)printf("}, {");						

					

					break;
				

		  	 	}//Ende case 30 power (phase 2)




	 		}//Ende switch(value)
		
			break;
		}//Ende Case 2 (Phase 2)









		//Phase 3
		case 3:
		{
			
			printf("\"phase\": %i,",phase_write);
			printf("\"name\": \"phase %i\",",phase_write);
			printf("\"values\": [{");
 		
			switch (value)
			{
				//Current:
				case 30:
				{
										
						summe_effektivwert=0;
						strom_klemme=0;
						ergebnis=0;
						effektivwert=0;
						strom_real=0;

 						for (z=0; z<500; z++) //In dieser Schleife werden 500 Strommessungen durchgeführt und anschließend der Mittelwert berechnet!!!
 						{

  							Daten[0] = 0x43;//0x43C0 (AIRMS; Current rms an A)
  				
							if (phase_write==1)	   Daten[1] = 0xC0;//Current Phase A
							else if (phase_write==2)Daten[1] = 0xC2;//Current Phase B
							else if (phase_write==3)Daten[1] = 0xC4;//Current Phase C
							else if (phase_write==4)Daten[1] = 0xC6;//Current N

  	 	 					if(write(device, Daten, 2) != 2)
  	 	 					{
  	 			 				//printf("Fehler beim Schreiben der Daten!\n");
  	  			 				return -1;
 	 	 					}
  	  
  	  						if (read(device, Daten, BUFSIZE) != BUFSIZE)
  	  						{
  	 	 	 					//printf("Fehler beim Lesen der Daten!\n");
 	  	  	 					return -1;
							}
 	 
  	 						ergebnis=(faktor_3*Daten[0])+(faktor_2*Daten[1])+(faktor_1*Daten[2])+Daten[3];   
       
 	 						effektivwert=(ergebnis*0.3535)/(rms_faktor_current); 	 

        						summe_effektivwert=summe_effektivwert+effektivwert;//Hier werden die 500 Messungen aufsummiert!!!
						}

  	 					effektivwert=summe_effektivwert/500; //Hier wird der Mittelwert von 500 Messungen berechnet!!!	
    
   	 					strom_klemme=effektivwert/7.07107;//10 Ohm ist der Widerstand der Stromklemme, parallel mit Widerstand (7,07107 Ohm)
  
   	 					offset_current=0.97129167;//OFFSET HABE ICH HIER MAL BEI 5A ermittelt!!!   

   	 					strom_real=(strom_klemme/0.05)*100*offset_current; //OFFSET HABE ICH HIER MAL BEI 5A ermittelt!!!

	 					//printf("\r\nStrom Phase:%i (effektiv): %f A",phase_write,strom_real);
					
						printf("\"type\": \"current\",");
						printf("\"unity\": \"A\",");
						printf("\"data\": %f",strom_real);
						printf("}, {");



						spannung_real=0;
						ergebnis=0;
						summe_ergebnis=0;

 						for (z=0; z<500; z++) //In dieser Schleife werden 500 Strommessungen durchgeführt und anschließend der Mittelwert berechnet!!!
 						{

  							Daten[0] = 0x43;//0x43C5 (CVRMS; Spannung rms an VC)
  							Daten[1] = 0xC5; 
 

  							if(write(device, Daten, 2) != 2)
  							{
  	  							//printf("Fehler beim Schreiben der Daten!\n");
  	 							return -1;
 							}

      


  	 
  	 						if (read(device, Daten, BUFSIZE) != BUFSIZE)
  	 						{
  	  							//printf("Fehler beim Lesen der Daten!\n");
 	 							return -1;
 	 						} 
	
  	 						ergebnis=(faktor_3*Daten[0])+(faktor_2*Daten[1])+(faktor_1*Daten[2])+Daten[3];    	
       

        						summe_ergebnis=summe_ergebnis+ergebnis;

							//Ich mache es hier anders als bei der Strommessung.
							//Ich habe bei 229,8 V (eff.) einen Wert von 2427873 erhalten
							//Ich summiere hier einfach 500 mal die gemessenen Werte auf und ermittle anschließend die Spannung!
						}

  	 					ergebnis=summe_ergebnis/500; //Hier wird der Mittelwert von 500 Messungen berechnet!!! 
  						//2427873 entspricht ca. 229.8 V, somit hat man bei maximalen 4191910 einen maximal zu messenen Wert von ca. 400 V (EFFEKTIV) 
   						spannung_real=(ergebnis/rms_faktor_voltage)*229.8*offset_voltage; //OFFSET HABE ICH HIER ERSTMAL AUF 1 gesetzt!!! rms_faktor ist hier 2427873 und nicht 4191910!					
						printf("\"type\": \"voltage\",");
						printf("\"unity\": \"V\",");
						printf("\"data\": %f",spannung_real);
						printf("}, {");

						//Power:
						power=strom_real*spannung_real;
						printf("\"type\": \"power\",");
						printf("\"unity\": \"W\",");
						printf("\"data\": %f",power);
												

					

					break;
				

		  	 	}//Ende case 30 power (phase 3)




	 		}//Ende switch(value)
		
			break;
		}//Ende Case 3 (Phase 3)











		


		



		default:
  		{
   	 		//Wenn keine Wahl zutrifft!!!
   	 		break;
  		}



	}//Ende der Switch-Anweisung (PHASE)

}//Ende der for-Schleife
printf("}]");
printf("}]");
printf("}]");
printf("}");

  close(device);//I2C-Verbindung schließen!!!
  return 0;
}